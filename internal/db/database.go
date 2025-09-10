package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Database represents the database connection pool
type Database struct {
	Pool *pgxpool.Pool
}

// NewDatabase creates a new database connection pool
func NewDatabase() (*Database, error) {
	// Load environment variables
	err := loadEnv()
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
		log.Println("Falling back to system environment variables")
	}

	// Database connection string from ENV variable
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL is empty and is required")
	}

	// Parse configuration
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse configuration: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = 50
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 5 * time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create the connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("✅ Successfully connected to the database")
	return &Database{Pool: pool}, nil
}

// loadEnv loads environment variables from .env file
func loadEnv() error {
	// Try to get the project root directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("could not get current file path")
	}

	// Navigate up to project root (adjust based on your structure)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	envPath := filepath.Join(projectRoot, ".env")

	return godotenv.Load(envPath)
}

// Close closes the database connection pool
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("✅ Database connection pool closed")
	}
}

// HealthCheck verifies the database connection is still alive
func (db *Database) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.Pool.Ping(ctx)
}

// GetPool returns the connection pool
func (db *Database) GetPool() *pgxpool.Pool {
	return db.Pool
}
