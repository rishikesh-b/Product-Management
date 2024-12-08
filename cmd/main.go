package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"product-management/internal/api"
	"product-management/internal/cache"
	"product-management/internal/db"
	"product-management/internal/logging"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Initialize logging
	logging.Init()

	// Connect to the database
	if err := db.Connect(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Connect to Redis
	if err := cache.Connect(); err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}

	// Initialize the router
	router := mux.NewRouter()

	// Define the API routes
	router.HandleFunc("/products", api.CreateProduct).Methods("POST")
	router.HandleFunc("/products/{id:[0-9]+}", api.GetProductByID).Methods("GET")
	router.HandleFunc("/products", api.GetProducts).Methods("GET")
	router.HandleFunc("/products/{id:[0-9]+}", api.UpdateProduct).Methods("PUT")

	// Setup the server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second, // Max duration before the server closes a read
		WriteTimeout: 10 * time.Second, // Max duration before the server closes a write
	}

	// Graceful shutdown using OS signal handling (Ctrl+C)
	// Use errgroup to manage graceful shutdown in case of errors
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		// Listen and serve
		log.Println("Server is running on port 8080")
		return server.ListenAndServe()
	})

	// Handle OS interrupts (SIGINT)
	g.Go(func() error {
		// Wait for interrupt signal (e.g., Ctrl+C)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)

		// Wait for interrupt signal
		<-sigCh
		log.Println("Received shutdown signal, shutting down gracefully...")

		// Graceful shutdown: Cancel any ongoing requests and stop the server
		cancelTimeout := 5 * time.Second // Timeout duration for gracefully shutting down
		ctx, cancel := context.WithTimeout(ctx, cancelTimeout)
		defer cancel()

		// Graceful shutdown logic
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Error during graceful shutdown: %v", err)
		}
		log.Println("Server stopped gracefully")
		return nil
	})

	// Wait for either the server or the shutdown signal to finish
	if err := g.Wait(); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}
