// File: internal/handlers/handlers.go

package handlers

import (
    "encoding/csv"
    "encoding/json"
    "errors"
    "html/template"
    "net/http"
    "fmt"

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
    tmpl, err := template.ParseGlob("web/templates/*.html")
    if err != nil {
        return nil, errors.Join(errors.New("failed to parse templates"), err)
    }

    return &Handler{
        db:        db,
        templates: tmpl,
    }, nil
}

// HomeHandler displays the main page
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {

    // Get statistics with context
    stats, err := h.db.GetStats()
    if err != nil {
        http.Error(w, "Error fetching statistics", http.StatusInternalServerError)
        return
    }

    // Render template
    if err := h.templates.ExecuteTemplate(w, "index.html", stats); err != nil {
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

    // Save to database with context
    if err := h.db.SaveReading(avg); err != nil {
        respondWithError(w, "Error saving reading", http.StatusInternalServerError)
        return
    }

    // Get updated statistics
    stats, err := h.db.GetStats()
    if err != nil {
        respondWithError(w, "Error fetching statistics", http.StatusInternalServerError)
        return
    }

    // Return success response
    response := map[string]interface{}{
        "message": "Reading saved successfully",
        "stats":   stats,
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
        http.Error(w, "Error writing CSV headers", http.StatusInternalServerError)
        return
    }

    readings, err := h.db.GetAllReadings()
    if err != nil {
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
            http.Error(w, "Error writing CSV data", http.StatusInternalServerError)
            return
        }
    }
}

// respondWithError sends an error response as JSON
func respondWithError(w http.ResponseWriter, message string, code int) {
    response := map[string]string{"error": message}
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
    }
}

// respondWithJSON sends a success response as JSON
func respondWithJSON(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(data); err != nil {
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
    }
}

// Middleware to add basic security headers
func (h *Handler) securityHeaders(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        next(w, r)
    }
}

// Routes returns the handler routes with middleware applied
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
