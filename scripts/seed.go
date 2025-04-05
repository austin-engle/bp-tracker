// File: scripts/seed.go

package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"bp-tracker/internal/database"
	"bp-tracker/internal/models"
	"bp-tracker/internal/utils"
)

func main() {
    // Command line flags
    dbPath := flag.String("db", "bp.db", "Path to SQLite database")
    days := flag.Int("days", 60, "Number of days of data to generate")
    flag.Parse()

    // Initialize database
    db, err := database.New(*dbPath)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    // Generate readings for each day
    startDate := time.Now().AddDate(0, 0, -(*days))
    for day := 0; day < *days; day++ {
        // Create 1-3 readings per day
        numReadings := rand.Intn(3) + 1

        for i := 0; i < numReadings; i++ {
            // Generate realistic blood pressure values
            systolic := 110 + rand.Intn(40)  // 110-150 range
            diastolic := 70 + rand.Intn(20)  // 70-90 range
            pulse := 60 + rand.Intn(30)      // 60-90 range

            // Create reading
            reading := &models.Reading{
                Timestamp:  startDate.AddDate(0, 0, day).Add(time.Duration(i*4) * time.Hour),
                Systolic:   systolic,
                Diastolic:  diastolic,
                Pulse:      pulse,
                Classification: utils.ClassifyBP(systolic, diastolic).Name,
            }

            // Save to database
            if err := db.SaveReading(reading); err != nil {
                log.Printf("Error saving reading: %v", err)
            }
        }
    }

    log.Printf("Successfully generated %d days of readings", *days)
}
