// File: internal/database/db.go

package database

import (
	"context"
	"database/sql" // Added for secret parsing
	"fmt"
	"log" // Added for logging
	"os"
	"time"

	"bp-tracker/internal/models"

	"github.com/aws/aws-sdk-go-v2/config"                 // Added AWS SDK config
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager" // Added Secrets Manager client
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB wraps the SQL database connection
type DB struct {
	*sql.DB
}

// Function to get secret from AWS Secrets Manager
func getSecret(secretARN string) (string, error) {
	ctx := context.TODO() // Use TODO context for this utility function
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	svc := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretARN,
	}

	result, err := svc.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get secret value: %w", err)
	}

	// Check if the secret string is nil first
	if result.SecretString == nil {
		// Handle binary secret if necessary, or return an error if string is expected
		// For now, let's assume a nil string is an error for a password.
		return "", fmt.Errorf("secret string is nil for ARN: %s", *input.SecretId) // Use input.SecretId for ARN in error
	}

	// Return the raw secret string directly, assuming it's the password
	return *result.SecretString, nil
}

// New creates a new PostgreSQL database connection using environment variables and Secrets Manager
func New() (*DB, error) {
	// Read connection details from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	secretARN := os.Getenv("SECRET_ARN") // Read Secret ARN
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	// Basic validation
	if dbHost == "" || dbPort == "" || dbUser == "" || secretARN == "" || dbName == "" || dbSSLMode == "" {
		return nil, fmt.Errorf("missing required environment variables (DB_HOST, DB_PORT, DB_USER, SECRET_ARN, DB_NAME, DB_SSLMODE)")
	}

	// Fetch the password from Secrets Manager
	dbPassword, err := getSecret(secretARN)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DB password from Secrets Manager (ARN: %s): %w", secretARN, err)
	}

	// Construct the DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	// Use "pgx" as the driver name with stdlib
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection pool: %w", err)
	}

	// Test the connection using context for potential timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Slightly longer timeout
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		// Provide more context on ping failure
		db.Close() // Attempt to close DB on failure
		return nil, fmt.Errorf("error pinging database (host=%s, user=%s, db=%s): %w", dbHost, dbUser, dbName, err)
	}

	log.Println("Successfully connected to PostgreSQL database!") // Use log package

	return &DB{db}, nil
}

// SaveReading stores a new blood pressure reading using PostgreSQL syntax
func (db *DB) SaveReading(r *models.Reading) error {
	query := `
        INSERT INTO readings (timestamp, systolic, diastolic, pulse, classification)
        VALUES ($1, $2, $3, $4, $5)
    ` // Changed placeholders, removed strftime

	// Pass the time.Time directly, pgx handles it
	_, err := db.Exec(query, r.Timestamp, r.Systolic, r.Diastolic, r.Pulse, r.Classification)
	if err != nil {
		return fmt.Errorf("error saving reading: %w", err)
	}

	return nil
}

// GetStats retrieves blood pressure statistics using PostgreSQL syntax
func (db *DB) GetStats() (*models.Stats, error) {
	stats := &models.Stats{}

	// Get last reading - select timestamp directly, use $ placeholders if needed (none here)
	lastReadingQuery := `
        SELECT id, timestamp as ts, systolic, diastolic, pulse, classification
        FROM readings
        ORDER BY timestamp DESC
        LIMIT 1
    `

	stats.LastReading = &models.Reading{}
	// Scan directly into time.Time
	err := db.QueryRow(lastReadingQuery).Scan(
		&stats.LastReading.ID,
		&stats.LastReading.Timestamp, // Scan directly into time.Time
		&stats.LastReading.Systolic,
		&stats.LastReading.Diastolic,
		&stats.LastReading.Pulse,
		&stats.LastReading.Classification,
	)
	if err == sql.ErrNoRows {
		stats.LastReading = nil
		return stats, nil // Return empty stats if no data
	} else if err != nil {
		return nil, fmt.Errorf("error getting last reading: %w", err)
	}

	// Time ranges for averages
	now := time.Now() // Use standard time.Now() unless timezone logic is critical
	sevenDaysAgo := now.AddDate(0, 0, -7)
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Helper function for getting averages within a specific time range using PostgreSQL syntax
	getAverageWithRange := func(start, end time.Time) (*models.Reading, int, error) {
		query := `
            SELECT
                COALESCE(ROUND(AVG(systolic)), 0)::int as avg_systolic,
                COALESCE(ROUND(AVG(diastolic)), 0)::int as avg_diastolic,
                COALESCE(ROUND(AVG(pulse)), 0)::int as avg_pulse,
                COUNT(*) as reading_count
            FROM readings
            WHERE timestamp >= $1 AND timestamp < $2
        ` // Changed placeholders, timestamp comparison, added ::int cast for Scan

		r := &models.Reading{}
		var count int
		// Pass time.Time directly
		err := db.QueryRow(query, start, end).Scan(&r.Systolic, &r.Diastolic, &r.Pulse, &count)
		if err != nil {
			if err == sql.ErrNoRows && count == 0 {
				return nil, 0, nil
			}
			return nil, 0, fmt.Errorf("error executing average query for range %v to %v: %w", start, end, err)
		}
		if count == 0 {
			return nil, 0, nil
		}

		return r, count, nil
	}

	// Get averages for different time periods
	stats.SevenDayAvg, stats.SevenDayCount, err = getAverageWithRange(sevenDaysAgo, now)
	if err != nil {
		log.Printf("Error getting 7-day average: %v\n", err)
		return nil, fmt.Errorf("error calculating 7-day average: %w", err)
	}

	stats.ThirtyDayAvg, stats.ThirtyDayCount, err = getAverageWithRange(thirtyDaysAgo, now)
	if err != nil {
		log.Printf("Error getting 30-day average: %v\n", err)
		return nil, fmt.Errorf("error calculating 30-day average: %w", err)
	}

	// All time average using PostgreSQL syntax
	allTimeQuery := `
        SELECT
            COALESCE(ROUND(AVG(systolic)), 0)::int as avg_systolic,
            COALESCE(ROUND(AVG(diastolic)), 0)::int as avg_diastolic,
            COALESCE(ROUND(AVG(pulse)), 0)::int as avg_pulse,
            COUNT(*) as reading_count
        FROM readings
    `
	r := &models.Reading{}
	var count int
	err = db.QueryRow(allTimeQuery).Scan(&r.Systolic, &r.Diastolic, &r.Pulse, &count)
	if err != nil {
		if err == sql.ErrNoRows && count == 0 {
			// No data, do nothing
		} else {
			return nil, fmt.Errorf("error getting all-time average: %w", err)
		}
	}
	if count > 0 {
		stats.AllTimeAvg = r
		stats.AllTimeCount = count
	}

	return stats, nil
}

// GetAllReadings retrieves all readings using PostgreSQL syntax
func (db *DB) GetAllReadings() ([]*models.Reading, error) {
	query := `
        SELECT id, timestamp as ts, systolic, diastolic, pulse, classification
        FROM readings
        ORDER BY timestamp DESC
    ` // Removed datetime(), select timestamp directly

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying readings: %w", err)
	}
	defer rows.Close()

	var readings []*models.Reading
	for rows.Next() {
		r := &models.Reading{}
		// Scan directly into time.Time
		err := rows.Scan(&r.ID, &r.Timestamp, &r.Systolic, &r.Diastolic, &r.Pulse, &r.Classification)
		if err != nil {
			return nil, fmt.Errorf("error scanning reading: %w", err)
		}
		readings = append(readings, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating readings: %w", err)
	}

	return readings, nil
}

// ClearAllReadings deletes all entries from the readings table.
// WARNING: Use with caution, typically only for testing/development.
func (db *DB) ClearAllReadings() error {
	query := `DELETE FROM readings`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error clearing readings table: %w", err)
	}
	log.Println("Successfully cleared all readings from the database.")
	return nil
}

// SeedReadings inserts multiple sample readings into the database.
// Uses a transaction for efficiency and atomicity.
func (db *DB) SeedReadings(readings []*models.Reading) error {
	ctx := context.Background() // Use background context for seeding operation
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction for seeding: %w", err)
	}
	// Defer rollback in case of error
	defer tx.Rollback() // Rollback is a no-op if Commit succeeds

	query := `
        INSERT INTO readings (timestamp, systolic, diastolic, pulse, classification)
        VALUES ($1, $2, $3, $4, $5)
    `
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing statement for seeding: %w", err)
	}
	defer stmt.Close()

	for _, r := range readings {
		_, err := stmt.ExecContext(ctx, r.Timestamp, r.Systolic, r.Diastolic, r.Pulse, r.Classification)
		if err != nil {
			// Error occurred, transaction will be rolled back by defer
			return fmt.Errorf("error inserting seed reading (timestamp %v): %w", r.Timestamp, err)
		}
	}

	// All insertions successful, commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction for seeding: %w", err)
	}

	log.Printf("Successfully seeded %d readings into the database.\n", len(readings))
	return nil
}

// DeleteReading deletes a specific reading by its ID.
func (db *DB) DeleteReading(id int64) error {
	query := `DELETE FROM readings WHERE id = $1`
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error executing delete query for id %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Log error but don't necessarily fail the operation if delete worked
		log.Printf("Warning: Could not get rows affected after delete for id %d: %v", id, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no reading found with id %d to delete", id)
	}

	log.Printf("Successfully deleted reading with id %d (%d rows affected)\n", id, rowsAffected)
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
