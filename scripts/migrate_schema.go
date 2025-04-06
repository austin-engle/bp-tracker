// File: scripts/migrate_schema.go

package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

func main() {
	fmt.Println("Starting schema migration...")

	// --- Database Connection (mirrors internal/database/db.go New() logic) ---
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" || dbSSLMode == "" {
		fmt.Fprintf(os.Stderr, "Error: Missing required environment variables (DB_HOST, DB_PORT, DB_USER, DB_NAME, DB_SSLMODE)\n")
		fmt.Fprintf(os.Stderr, "DB_PASSWORD may also be required depending on auth.\n")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database connection: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Increased timeout for potential network latency
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully connected to database %s@%s:%s\n", dbUser, dbHost, dbPort)

	// --- Read Schema File ---
	// Assumes script is run from the project root (bp-tracker/)
	schemaPath := filepath.Join("internal", "database", "schema.sql")
	schemaBytes, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading schema file %s: %v\n", schemaPath, err)
		os.Exit(1)
	}
	schemaSQL := string(schemaBytes)

	// --- Execute Schema ---
	// Using a transaction is good practice, although this simple schema might not strictly need it.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting transaction: %v\n", err)
		os.Exit(1)
	}

	_, err = tx.ExecContext(ctx, schemaSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing schema: %v\n", err)
		// Attempt to rollback
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			fmt.Fprintf(os.Stderr, "Error rolling back transaction: %v\n", rollbackErr)
		}
		os.Exit(1)
	}

	if err := tx.Commit(); err != nil {
		fmt.Fprintf(os.Stderr, "Error committing transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Schema migration applied successfully!")
}
