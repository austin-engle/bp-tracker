// File: internal/handlers/handlers.go

package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bp-tracker/internal/database"
	"bp-tracker/internal/models"
	"bp-tracker/internal/utils"
	"bp-tracker/internal/validation"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db        *database.DB
	templates *template.Template
}

// New creates a new Handler instance
func New(db *database.DB) (*Handler, error) {
	log.Println("Parsing templates...")
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse templates"), err)
	}
	log.Println("Templates parsed successfully.")

	h := &Handler{
		db:        db,
		templates: tmpl,
	}

	// Log handler methods to confirm presence
	log.Printf("Handler created. HomeHandler: %p, MigrateHandler: %p\n", h.HomeHandler, h.MigrateHandler)

	return h, nil
}

// HomeHandler displays the main page
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {

	// Get statistics with context
	stats, err := h.db.GetStats()
	if err != nil {
		log.Printf("ERROR HomeHandler - fetching stats: %v", err)
		http.Error(w, "Error fetching statistics", http.StatusInternalServerError)
		return
	}

	// Render template
	if err := h.templates.ExecuteTemplate(w, "index.html", stats); err != nil {
		log.Printf("ERROR HomeHandler - rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// SubmitReadingHandler processes new blood pressure readings
func (h *Handler) SubmitReadingHandler(w http.ResponseWriter, r *http.Request) {

	var input models.ReadingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, "Invalid input format", http.StatusBadRequest)
		return
	}

	// Validate readings
	if err := validation.ValidateReadings(&input); err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate average
	avg := input.Average()

	// Classify blood pressure
	category := utils.ClassifyBP(avg.Systolic, avg.Diastolic)
	avg.Classification = category.Name
	avg.Timestamp = time.Now() // Ensure timestamp is set

	// Save to database with context
	if err := h.db.SaveReading(avg); err != nil {
		log.Printf("ERROR SubmitReadingHandler - saving reading: %v", err)
		respondWithError(w, "Error saving reading", http.StatusInternalServerError)
		return
	}

	// Get updated statistics
	stats, err := h.db.GetStats()
	if err != nil {
		// Log error but maybe still return success? Or return error?
		log.Printf("ERROR SubmitReadingHandler - fetching stats after save: %v", err)
		respondWithError(w, "Error fetching statistics after save", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message":        "Reading saved successfully",
		"stats":          stats,
		"classification": category,
		"recommendation": utils.GetRecommendation(category),
	}

	respondWithJSON(w, response)
}

// ExportCSVHandler handles the export of readings to CSV
func (h *Handler) ExportCSVHandler(w http.ResponseWriter, r *http.Request) {

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=blood_pressure_readings.csv")

	// Create CSV writer
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	headers := []string{"Date", "Time", "Systolic", "Diastolic", "Pulse", "Classification"}
	if err := writer.Write(headers); err != nil {
		log.Printf("ERROR ExportCSVHandler - writing header: %v", err)
		http.Error(w, "Error writing CSV headers", http.StatusInternalServerError)
		return
	}

	readings, err := h.db.GetAllReadings()
	if err != nil {
		log.Printf("ERROR ExportCSVHandler - fetching readings: %v", err)
		http.Error(w, "Error fetching readings", http.StatusInternalServerError)
		return
	}

	// Write readings to CSV
	for _, reading := range readings {
		record := []string{
			reading.Timestamp.Format("2006-01-02"),
			reading.Timestamp.Format("15:04:05"),
			fmt.Sprintf("%d", reading.Systolic),
			fmt.Sprintf("%d", reading.Diastolic),
			fmt.Sprintf("%d", reading.Pulse),
			reading.Classification,
		}

		if err := writer.Write(record); err != nil {
			log.Printf("ERROR ExportCSVHandler - writing record: %v", err)
			http.Error(w, "Error writing CSV data", http.StatusInternalServerError)
			return // Stop writing if one record fails
		}
	}
}

// --- NEW HANDLER ---
// MigrateHandler applies the schema.sql file to the database.
// WARNING: This endpoint should be secured, ideally via IAM authorization.
func (h *Handler) MigrateHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for /migrate")

	// Define path relative to where binary runs in Lambda (/var/task)
	schemaPath := "schema.sql"
	schemaBytes, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		// Specific error for file reading
		msg := fmt.Sprintf("MIGRATION_ERROR: Error reading schema file %s: %v", schemaPath, err)
		log.Println(msg) // Use Println for consistency
		respondWithError(w, msg, http.StatusInternalServerError)
		return
	}
	schemaSQL := string(schemaBytes)

	// Execute schema in a transaction
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second) // Use request context with timeout
	defer cancel()

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		// Specific error for transaction start
		msg := fmt.Sprintf("MIGRATION_ERROR: Error starting transaction: %v", err)
		log.Println(msg)
		respondWithError(w, msg, http.StatusInternalServerError)
		return
	}

	_, err = tx.ExecContext(ctx, schemaSQL)
	if err != nil {
		// Specific error for schema execution
		msg := fmt.Sprintf("MIGRATION_ERROR: Error executing schema: %v", err)
		log.Println(msg)
		// Attempt to rollback
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("ERROR MigrateHandler - Error rolling back transaction after schema execution failure: %v", rollbackErr)
		}
		respondWithError(w, msg, http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		// Specific error for commit
		msg := fmt.Sprintf("MIGRATION_ERROR: Error committing transaction: %v", err)
		log.Println(msg)
		respondWithError(w, msg, http.StatusInternalServerError)
		return
	}

	log.Println("Schema migration applied successfully via /migrate endpoint.")
	respondWithJSON(w, map[string]string{"message": "Schema migration applied successfully!"})
}

// --- NEW HANDLER ---
// GetAllReadingsJSONHandler handles the request to fetch all readings as JSON.
func (h *Handler) GetAllReadingsJSONHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for /api/readings")

	readings, err := h.db.GetAllReadings()
	if err != nil {
		log.Printf("ERROR GetAllReadingsJSONHandler - fetching readings: %v", err)
		respondWithError(w, "Error fetching readings", http.StatusInternalServerError)
		return
	}

	// Handle case where there are no readings
	if readings == nil {
		readings = []*models.Reading{} // Return empty slice of pointers
	}

	log.Printf("Successfully fetched %d readings for /api/readings", len(readings))
	respondWithJSON(w, readings)
}

// --- NEW HANDLER ---
// GetStatsHandler handles the request to fetch statistics as JSON.
func (h *Handler) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for /api/stats")

	stats, err := h.db.GetStats()
	if err != nil {
		log.Printf("ERROR GetStatsHandler - fetching stats: %v", err)
		respondWithError(w, "Error fetching statistics", http.StatusInternalServerError)
		return
	}

	// Handle case where stats might be partially nil (e.g., no readings yet)
	// The frontend should handle nil values gracefully.
	if stats == nil {
		// Respond with an empty object or default stats if preferred?
		// For now, let's rely on the DB layer returning a valid (potentially zeroed) Stats object.
		// If h.db.GetStats can return nil, we might need:
		// stats = &models.Stats{}
		log.Println("Warning: GetStats returned nil, sending potentially empty stats object.")
	}

	log.Println("Successfully fetched stats for /api/stats")
	respondWithJSON(w, stats)
}

// --- DEVELOPMENT HANDLERS ---

// SeedHandler adds sample data to the database.
func (h *Handler) SeedHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for /api/dev/seed")

	// Define sample readings (adjust as needed)
	now := time.Now()
	seedData := []*models.Reading{
		// > 90 days ago
		{Timestamp: now.AddDate(0, 0, -100), Systolic: 125, Diastolic: 83, Pulse: 70, Classification: "Hypertension Stage 1"},
		{Timestamp: now.AddDate(0, 0, -95), Systolic: 120, Diastolic: 79, Pulse: 68, Classification: "Normal"},

		// ~ 30-90 days ago
		{Timestamp: now.AddDate(0, 0, -45), Systolic: 133, Diastolic: 86, Pulse: 74, Classification: "Hypertension Stage 1"},
		{Timestamp: now.AddDate(0, 0, -35), Systolic: 128, Diastolic: 82, Pulse: 71, Classification: "Elevated"},

		// ~ 7-30 days ago
		{Timestamp: now.AddDate(0, 0, -25), Systolic: 142, Diastolic: 91, Pulse: 78, Classification: "Hypertension Stage 2"},
		{Timestamp: now.AddDate(0, 0, -15), Systolic: 138, Diastolic: 87, Pulse: 76, Classification: "Hypertension Stage 1"},
		{Timestamp: now.AddDate(0, 0, -10), Systolic: 126, Diastolic: 83, Pulse: 72, Classification: "Hypertension Stage 1"},
		{Timestamp: now.AddDate(0, 0, -8), Systolic: 121, Diastolic: 79, Pulse: 69, Classification: "Normal"},

		// < 7 days ago
		{Timestamp: now.AddDate(0, 0, -6), Systolic: 118, Diastolic: 78, Pulse: 65, Classification: "Normal"},
		{Timestamp: now.AddDate(0, 0, -5), Systolic: 122, Diastolic: 81, Pulse: 70, Classification: "Elevated"},
		{Timestamp: now.AddDate(0, 0, -3), Systolic: 135, Diastolic: 88, Pulse: 75, Classification: "Hypertension Stage 1"},
		{Timestamp: now.AddDate(0, 0, -1), Systolic: 141, Diastolic: 90, Pulse: 76, Classification: "Hypertension Stage 2"},
	}

	err := h.db.SeedReadings(seedData)
	if err != nil {
		log.Printf("ERROR SeedHandler: %v", err)
		respondWithError(w, fmt.Sprintf("Error seeding data: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": fmt.Sprintf("Successfully seeded %d readings", len(seedData))})
}

// ClearHandler removes all data from the readings table.
func (h *Handler) ClearHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for /api/dev/clear")

	err := h.db.ClearAllReadings()
	if err != nil {
		log.Printf("ERROR ClearHandler: %v", err)
		respondWithError(w, fmt.Sprintf("Error clearing data: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Successfully cleared all readings"})
}

// DeleteReadingHandler handles requests to delete a specific reading.
func (h *Handler) DeleteReadingHandler(w http.ResponseWriter, r *http.Request) {
	// Need to get the ID from the URL path parameter
	// Using ginadapter means we access parameters via gin context,
	// but since we wrapped http.HandlerFunc, we need a way to access it.
	// This might require adjusting how routes are handled or using a different adapter
	// if direct path param access is needed without full Gin context.

	// --- WORKAROUND (Less Ideal): Extract ID from URL path string ---
	// Assumes URL like /api/readings/123
	// NOTE: This is fragile and depends on the exact path registered in main.go
	// A better solution involves using a router that properly populates context/vars.
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 { // Expecting ["api", "readings", "id"]
		 log.Printf("ERROR DeleteReadingHandler: Invalid path structure: %s", r.URL.Path)
		 respondWithError(w, "Invalid request path for deleting reading", http.StatusBadRequest)
		 return
	}
	idStr := pathParts[len(pathParts)-1] // Get the last part as ID
	// --- End Workaround ---

	log.Printf("Received request to delete reading with ID string: %s", idStr)

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("ERROR DeleteReadingHandler: Invalid ID format '%s': %v", idStr, err)
		respondWithError(w, "Invalid reading ID format", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteReading(id)
	if err != nil {
		// Check if it's a "not found" error specifically
		if strings.Contains(err.Error(), "no reading found") {
			 log.Printf("INFO DeleteReadingHandler: Reading ID %d not found: %v", id, err)
			 respondWithError(w, err.Error(), http.StatusNotFound) // 404 Not Found
		} else {
			 log.Printf("ERROR DeleteReadingHandler: Failed to delete ID %d: %v", id, err)
			respondWithError(w, fmt.Sprintf("Error deleting reading: %v", err), http.StatusInternalServerError)
		}
		return
	}

	respondWithJSON(w, map[string]string{"message": fmt.Sprintf("Successfully deleted reading %d", id)})
}

// respondWithError sends an error response as JSON
func respondWithError(w http.ResponseWriter, message string, code int) {
	log.Printf("Responding with error (Code %d): %s", code, message) // Add logging here
	response := map[string]string{"error": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log error if response writing fails
		log.Printf("ERROR respondWithError - encoding response: %v", err)
	}
}

// respondWithJSON sends a success response as JSON
func respondWithJSON(w http.ResponseWriter, data interface{}) {
	log.Printf("Responding with success: %v", data) // Add logging here
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error if response writing fails
		log.Printf("ERROR respondWithJSON - encoding response: %v", err)
	}
}

// Removed securityHeaders middleware function as it's handled in main.go/Gin now
/*
func (h *Handler) securityHeaders(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        next(w, r)
    }
}
*/

// Removed Routes method as routing is now done in main.go/Gin
/*
func (h *Handler) Routes() *http.ServeMux {
    mux := http.NewServeMux()

    // Apply security headers to all routes
    mux.HandleFunc("GET /", h.securityHeaders(h.HomeHandler))
    mux.HandleFunc("POST /submit", h.securityHeaders(h.SubmitReadingHandler))
    mux.HandleFunc("GET /export/csv", h.securityHeaders(h.ExportCSVHandler))

    // Static files
    mux.Handle("GET /static/", http.StripPrefix("/static/",
        http.FileServer(http.Dir("web/static"))))

    return mux
}
*/
