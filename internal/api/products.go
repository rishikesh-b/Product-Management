package api

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"product-management/internal/cache"
	"product-management/internal/db"
	"product-management/internal/logging"
	"product-management/internal/queue"
	"strconv"
	"strings"
)

type Product struct {
	UserID                  int      `json:"user_id"`
	ProductName             string   `json:"product_name"`
	ProductDesc             string   `json:"product_description"`
	ProductImages           []string `json:"product_images"`
	ProductPrice            float64  `json:"product_price"`
	CompressedProductImages []string `json:"compressed_product_images"`
}

// RespondWithError provides a standardized error response
func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// CreateProduct creates a new product
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		logging.Logger.WithField("error", err.Error()).Error("Failed to decode request body")
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if product.UserID == 0 || product.ProductName == "" || len(product.ProductImages) == 0 || product.ProductPrice <= 0 {
		logging.Logger.Warn("Missing required fields or invalid data")
		RespondWithError(w, http.StatusBadRequest, "Missing required fields or invalid data")
		return
	}

	query := `
		INSERT INTO products (user_id, product_name, product_description, product_price, product_images)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var productID int
	err := db.DB.QueryRow(context.Background(), query,
		product.UserID,
		product.ProductName,
		product.ProductDesc,
		product.ProductPrice,
		product.ProductImages,
	).Scan(&productID)
	if err != nil {
		logging.Logger.WithField("error", err.Error()).Error("Failed to insert product into database")
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	imageProcessingTask := map[string]interface{}{
		"product_id": productID,
		"image_urls": product.ProductImages,
	}
	if err := queue.Publish(imageProcessingTask); err != nil {
		logging.Logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"product_id": productID,
			"image_urls": product.ProductImages,
		}).Error("Failed to enqueue image processing task")
		RespondWithError(w, http.StatusInternalServerError, "Failed to enqueue image processing task")
		return
	}

	logging.Logger.WithField("product_id", productID).Info("Product created successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Product created successfully",
		"product_id": productID,
	})
}

// GetProductByID fetches a product by ID
func GetProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productIDStr := vars["id"]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	cacheKey := "product:" + strconv.Itoa(productID)
	cachedProduct, err := cache.GetCache(cacheKey)
	if err == nil && cachedProduct != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedProduct))
		return
	}

	var product Product
	err = db.DB.QueryRow(context.Background(), `
		SELECT user_id, product_name, product_description, product_price, product_images, compressed_product_images
		FROM products WHERE id=$1`, productID).
		Scan(&product.UserID, &product.ProductName, &product.ProductDesc, &product.ProductPrice, &product.ProductImages, &product.CompressedProductImages)
	if err != nil {
		logging.Logger.WithField("product_id", productID).Warn("Product not found")
		RespondWithError(w, http.StatusNotFound, "Product not found")
		return
	}

	productData, _ := json.Marshal(product)
	cache.SetCache(cacheKey, string(productData), 10*60)

	w.Header().Set("Content-Type", "application/json")
	w.Write(productData)
}

// GetProducts fetches all products for a user
func GetProducts(w http.ResponseWriter, r *http.Request) {
	// Parse user_id query parameter
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	// Parse other query parameters
	minPriceStr := r.URL.Query().Get("min_price")
	maxPriceStr := r.URL.Query().Get("max_price")
	productName := r.URL.Query().Get("product_name")

	var minPrice, maxPrice float64
	var queryParts []string
	var queryArgs []interface{}
	var query string

	// Base query to filter by user_id
	query = `SELECT id, product_name, product_description, product_price, product_images, compressed_product_images FROM products WHERE user_id=$1`
	queryArgs = append(queryArgs, userID)
	queryParts = append(queryParts, "user_id=$1")

	// Check if min_price and max_price are provided
	if minPriceStr != "" {
		minPrice, err = strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid min_price")
			return
		}
		queryParts = append(queryParts, "product_price >= $2")
		queryArgs = append(queryArgs, minPrice)
	}

	if maxPriceStr != "" {
		maxPrice, err = strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid max_price")
			return
		}
		// Ensure that minPrice <= maxPrice
		if minPrice > maxPrice {
			RespondWithError(w, http.StatusBadRequest, "min_price should be less than or equal to max_price")
			return
		}
		queryParts = append(queryParts, "product_price <= $3")
		queryArgs = append(queryArgs, maxPrice)
	}

	// If a product name is provided, add it to the filter
	if productName != "" {
		queryParts = append(queryParts, "product_name ILIKE $4")
		queryArgs = append(queryArgs, "%"+productName+"%")
	}

	// Combine query parts using strings.Join
	if len(queryParts) > 1 {
		query += " AND " + strings.Join(queryParts[1:], " AND ")
	}

	// Debug: Log the query and parameters
	logging.Logger.WithFields(map[string]interface{}{
		"query":  query,
		"params": queryArgs,
	}).Info("Executing query to fetch products")

	// Execute the query
	rows, err := db.DB.Query(context.Background(), query, queryArgs...)
	if err != nil {
		logging.Logger.WithField("error", err.Error()).Error("Failed to retrieve products")
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve products")
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.UserID, &product.ProductName, &product.ProductDesc, &product.ProductPrice, &product.ProductImages, &product.CompressedProductImages); err != nil {
			logging.Logger.WithField("error", err.Error()).Error("Error scanning product data")
			RespondWithError(w, http.StatusInternalServerError, "Error scanning product data")
			return
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		logging.Logger.WithField("error", err.Error()).Error("Error iterating over products")
		RespondWithError(w, http.StatusInternalServerError, "Error iterating over products")
		return
	}

	// Cache the result
	productsData, _ := json.Marshal(products)
	cacheKey := "products:user:" + strconv.Itoa(userID)
	cache.SetCache(cacheKey, string(productsData), 10*60)

	// Respond with the filtered data
	w.Header().Set("Content-Type", "application/json")
	w.Write(productsData)
}

// UpdateProduct updates an existing product
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productIDStr := vars["id"]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		logging.Logger.WithField("error", err.Error()).Error("Failed to decode request body")
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if product.ProductName == "" || product.ProductPrice <= 0 {
		logging.Logger.Warn("Missing required fields or invalid data")
		RespondWithError(w, http.StatusBadRequest, "Missing required fields or invalid data")
		return
	}

	_, err = db.DB.Exec(context.Background(), `
		UPDATE products 
		SET product_name=$1, product_description=$2, product_price=$3, product_images=$4 
		WHERE id=$5`,
		product.ProductName,
		product.ProductDesc,
		product.ProductPrice,
		product.ProductImages,
		productID)
	if err != nil {
		logging.Logger.WithField("error", err.Error()).Error("Failed to update product")
		RespondWithError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}

	cacheKey := "product:" + strconv.Itoa(productID)
	cache.SetCache(cacheKey, "", 0) // Invalidate the cache

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Product updated successfully",
	})
}
