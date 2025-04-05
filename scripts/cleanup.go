// File: scripts/cleanup.go

package main

import (
	"flag"
	"log"
	"time"

	"bp-tracker/internal/database"
)

func main() {
    // Command line flags
    dbPath := flag.String("db", "bp.db", "Path to SQLite database")
    mode := flag.String("mode", "all", "Cleanup mode: all, before-date, after-date")
    date := flag.String("date", "", "Date for cleanup (format: YYYY-MM-DD)")
    flag.Parse()

    // Initialize database
    db, err := database.New(*dbPath)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    switch *mode {
    case "all":
        // Delete all readings
        if err := cleanAll(db); err != nil {
            log.Fatalf("Failed to clean database: %v", err)
        }
        log.Println("Successfully deleted all readings")

    case "before-date", "after-date":
        if *date == "" {
            log.Fatal("Date parameter is required for before-date/after-date modes")
        }

        // Parse date
        targetDate, err := time.Parse("2006-01-02", *date)
        if err != nil {
            log.Fatalf("Invalid date format. Use YYYY-MM-DD: %v", err)
        }

        // Clean based on date
        if err := cleanByDate(db, targetDate, *mode == "before-date"); err != nil {
            log.Fatalf("Failed to clean database: %v", err)
        }
        log.Printf("Successfully deleted readings %s %s",
            *mode, targetDate.Format("2006-01-02"))

    default:
        log.Fatalf("Invalid mode %q. Use: all, before-date, or after-date", *mode)
    }
}

// cleanAll removes all readings from the database
func cleanAll(db *database.DB) error {
    query := `DELETE FROM readings`
    _, err := db.Exec(query)
    return err
}

// cleanByDate removes readings before or after the specified date
func cleanByDate(db *database.DB, date time.Time, before bool) error {
    var query string
    if before {
        query = `DELETE FROM readings WHERE timestamp < ?`
    } else {
        query = `DELETE FROM readings WHERE timestamp > ?`
    }
    _, err := db.Exec(query, date.Unix())
    return err
}
