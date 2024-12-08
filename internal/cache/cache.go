package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"os"
	"product-management/internal/logging"
	"time"
)

var rdb *redis.Client

// Connect initializes the Redis client and establishes the connection.
func Connect() error {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("Error loading .env file: %w", err)
	}

	// Set Redis connection options from environment variables
	redisAddr := os.Getenv("REDIS_ADDR")
	redisDB := os.Getenv("REDIS_DB") // Default DB, if not specified

	if redisDB == "" {
		redisDB = "0" // Default to database 0
	}

	// Initialize Redis client without password
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr, // Redis server address (e.g., "localhost:6379")
		DB:   0,         // Default DB is 0
	})

	// Ping Redis to check if the connection is working
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		logging.Logger.Error("Error connecting to Redis: ", err)
		return fmt.Errorf("Error connecting to Redis: %w", err)
	}

	// Log successful Redis connection
	logging.Logger.Info("Connected to Redis successfully")

	return nil
}

// SetCache sets a key-value pair in Redis with an optional TTL (Time to Live)
func SetCache(key string, value string, ttl time.Duration) error {
	err := rdb.Set(context.Background(), key, value, ttl).Err()
	if err != nil {
		logging.Logger.Error("Error setting cache in Redis: ", err)
		return err
	}
	logging.Logger.WithFields(map[string]interface{}{
		"key": key,
		"ttl": ttl,
	}).Info("Cache set successfully in Redis")
	return nil
}

// GetCache retrieves the value of a key from Redis
func GetCache(key string) (string, error) {
	result, err := rdb.Get(context.Background(), key).Result()
	if err == redis.Nil {
		logging.Logger.Warn("Cache miss for key: ", key)
		return "", nil
	} else if err != nil {
		logging.Logger.Error("Error getting cache from Redis: ", err)
		return "", err
	}
	logging.Logger.WithFields(map[string]interface{}{
		"key": key,
	}).Info("Cache hit for key")
	return result, nil
}
