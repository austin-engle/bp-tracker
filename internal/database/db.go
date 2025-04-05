// File: internal/database/db.go

package database

import (
    "database/sql"
    "embed"
    "fmt"
    "time"

    "bp-tracker/internal/models"
    _ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps the SQL database connection
type DB struct {
    *sql.DB
}

// New creates a new database connection and initializes the schema
func New(dbPath string) (*DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("error opening database: %w", err)
    }

    // Initialize schema
    schema, err := schemaFS.ReadFile("schema.sql")
    if err != nil {
        return nil, fmt.Errorf("error reading schema: %w", err)
    }

    if _, err := db.Exec(string(schema)); err != nil {
        return nil, fmt.Errorf("error initializing schema: %w", err)
    }

    // Test the connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("error connecting to database: %w", err)
    }

    return &DB{db}, nil
}

// SaveReading stores a new blood pressure reading
func (db *DB) SaveReading(r *models.Reading) error {
    query := `
        INSERT INTO readings (timestamp, systolic, diastolic, pulse, classification)
        VALUES (strftime('%s', ?), ?, ?, ?, ?)
    `

    _, err := db.Exec(query, r.Timestamp.Format("2006-01-02 15:04:05"), r.Systolic, r.Diastolic, r.Pulse, r.Classification)
    if err != nil {
        return fmt.Errorf("error saving reading: %w", err)
    }

    return nil
}

// GetStats retrieves blood pressure statistics
func (db *DB) GetStats() (*models.Stats, error) {
    stats := &models.Stats{}

    // Get last reading
    lastReadingQuery := `
        SELECT id, datetime(timestamp, 'unixepoch') as ts, systolic, diastolic, pulse, classification
        FROM readings
        ORDER BY timestamp DESC
        LIMIT 1
    `

    stats.LastReading = &models.Reading{}
    var ts string
    err := db.QueryRow(lastReadingQuery).Scan(
        &stats.LastReading.ID,
        &ts,
        &stats.LastReading.Systolic,
        &stats.LastReading.Diastolic,
        &stats.LastReading.Pulse,
        &stats.LastReading.Classification,
    )
    if err == sql.ErrNoRows {
        stats.LastReading = nil
        // Return empty stats instead of error when no data exists
        return stats, nil
    } else if err != nil {
        return nil, fmt.Errorf("error getting last reading: %w", err)
    }

    // Parse the timestamp string into time.Time
    stats.LastReading.Timestamp, err = time.Parse("2006-01-02 15:04:05", ts)
    if err != nil {
        return nil, fmt.Errorf("error parsing timestamp: %w", err)
    }

    // Time ranges for averages
    now := models.GetTimestampInMST()
    sevenDaysAgo := now.AddDate(0, 0, -7)
    thirtyDaysAgo := now.AddDate(0, 0, -30)

    // Debug: Print all readings
    rows, err := db.Query(`
        SELECT datetime(timestamp, 'unixepoch') as ts, systolic, diastolic, pulse
        FROM readings
        ORDER BY timestamp DESC
    `)
    if err != nil {
        fmt.Printf("Error querying all readings: %v\n", err)
    } else {
        defer rows.Close()
        fmt.Println("\nAll readings in database:")
        for rows.Next() {
            var ts string
            var sys, dia, pul int
            if err := rows.Scan(&ts, &sys, &dia, &pul); err != nil {
                fmt.Printf("Error scanning row: %v\n", err)
                continue
            }
            fmt.Printf("%v: %d/%d, Pulse: %d\n", ts, sys, dia, pul)
        }
        fmt.Println()
    }

    fmt.Printf("Time ranges - Now: %v, 7 days ago: %v, 30 days ago: %v\n",
        now.Format("2006-01-02 15:04:05"),
        sevenDaysAgo.Format("2006-01-02 15:04:05"),
        thirtyDaysAgo.Format("2006-01-02 15:04:05"))

    // Helper function for getting averages within a specific time range
    getAverageWithRange := func(start, end time.Time) (*models.Reading, int, error) {
        query := `
            SELECT
                COALESCE(ROUND(AVG(systolic)), 0) as avg_systolic,
                COALESCE(ROUND(AVG(diastolic)), 0) as avg_diastolic,
                COALESCE(ROUND(AVG(pulse)), 0) as avg_pulse,
                COUNT(*) as reading_count
            FROM readings
            WHERE timestamp >= strftime('%s', ?) AND timestamp < strftime('%s', ?)
        `

        r := &models.Reading{}
        var count int
        err := db.QueryRow(query, start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05")).Scan(&r.Systolic, &r.Diastolic, &r.Pulse, &count)
        if err == sql.ErrNoRows || count == 0 {
            return nil, 0, nil
        } else if err != nil {
            return nil, 0, err
        }
        if r.Systolic == 0 && r.Diastolic == 0 && r.Pulse == 0 {
            return nil, 0, nil
        }
        fmt.Printf("Query from %v to %v returned: BP: %d/%d, Pulse: %d, Count: %d\n",
            start.Format("2006-01-02 15:04:05"),
            end.Format("2006-01-02 15:04:05"),
            r.Systolic, r.Diastolic, r.Pulse, count)
        return r, count, nil
    }

    // Get averages for different time periods
    // Last 7 days
    stats.SevenDayAvg, stats.SevenDayCount, err = getAverageWithRange(sevenDaysAgo, now)
    if err != nil {
        return nil, fmt.Errorf("error getting 7-day average: %w", err)
    }

    // Last 30 days
    stats.ThirtyDayAvg, stats.ThirtyDayCount, err = getAverageWithRange(thirtyDaysAgo, now)
    if err != nil {
        return nil, fmt.Errorf("error getting 30-day average: %w", err)
    }

    // All time
    allTimeQuery := `
        SELECT
            COALESCE(ROUND(AVG(systolic)), 0) as avg_systolic,
            COALESCE(ROUND(AVG(diastolic)), 0) as avg_diastolic,
            COALESCE(ROUND(AVG(pulse)), 0) as avg_pulse,
            COUNT(*) as reading_count
        FROM readings
    `
    r := &models.Reading{}
    var count int
    err = db.QueryRow(allTimeQuery).Scan(&r.Systolic, &r.Diastolic, &r.Pulse, &count)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("error getting all-time average: %w", err)
    }
    if count > 0 && (r.Systolic != 0 || r.Diastolic != 0 || r.Pulse != 0) {
        stats.AllTimeAvg = r
        stats.AllTimeCount = count
        fmt.Printf("All-time query returned: BP: %d/%d, Pulse: %d, Count: %d\n",
            r.Systolic, r.Diastolic, r.Pulse, count)
    }

    return stats, nil
}

// GetAllReadings retrieves all readings for CSV export
func (db *DB) GetAllReadings() ([]*models.Reading, error) {
    query := `
        SELECT id, datetime(timestamp, 'unixepoch') as ts, systolic, diastolic, pulse, classification
        FROM readings
        ORDER BY timestamp DESC
    `

    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("error querying readings: %w", err)
    }
    defer rows.Close()

    var readings []*models.Reading
    for rows.Next() {
        r := &models.Reading{}
        var ts string
        err := rows.Scan(&r.ID, &ts, &r.Systolic, &r.Diastolic, &r.Pulse, &r.Classification)
        if err != nil {
            return nil, fmt.Errorf("error scanning reading: %w", err)
        }
        r.Timestamp, err = time.Parse("2006-01-02 15:04:05", ts)
        if err != nil {
            return nil, fmt.Errorf("error parsing timestamp: %w", err)
        }
        readings = append(readings, r)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating readings: %w", err)
    }

    return readings, nil
}

// Close closes the database connection
func (db *DB) Close() error {
    return db.DB.Close()
}
