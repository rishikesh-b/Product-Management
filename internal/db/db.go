package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"os"
	"product-management/internal/logging"
)

var DB *pgxpool.Pool

// Connect initializes the database connection pool
func Connect() error {
	// Load environment variables from the .env file
	if err := godotenv.Load(); err != nil {
		logging.Logger.Error("Failed to load .env file", err)
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	// Get the database connection string from the environment variable
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logging.Logger.Error("DATABASE_URL environment variable is not set")
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// Connect to the database
	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		logging.Logger.Error("Unable to connect to database", err)
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Set the global DB connection pool
	DB = pool

	// Log successful connection
	logging.Logger.Info("Successfully connected to the database")

	return nil
}

// Close gracefully closes the database connection pool
func Close() {
	if DB != nil {
		DB.Close()
		logging.Logger.Info("Database connection pool closed successfully")
	}
}
