# Handlers Package

## Overview
The handlers package manages all HTTP endpoints for the blood pressure tracking application, utilizing Go 1.22's modern features for routing, context management, and error handling.

## Core Components

### Handler Structure
```go
type Handler struct {
    db        *database.DB
    templates *template.Template
}
```
- Dependency injection pattern
- Thread-safe shared resources
- Centralized template management

### Routes (Go 1.22+)
```go
func (h *Handler) Routes() *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /", h.securityHeaders(h.HomeHandler))
    mux.HandleFunc("POST /submit", h.securityHeaders(h.SubmitReadingHandler))
    mux.HandleFunc("GET /export/csv", h.securityHeaders(h.ExportCSVHandler))
    return mux
}
```

### 1. Home Route (`GET /`)
```go
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request)
```
- **Purpose**: Displays main page and current statistics
- **Method**: GET
- **Input**: None
- **Returns**: HTML page with:
  - Last reading details
  - 7-day averages
  - 30-day averages
  - All-time averages
- **Error Cases**:
  - Database errors (500)
  - Template rendering errors (500)
- **Template Used**: index.html

### 2. Submit Reading (`POST /submit`)
```go
func (h *Handler) SubmitReadingHandler(w http.ResponseWriter, r *http.Request)
```
- **Purpose**: Processes new blood pressure readings
- **Method**: POST
- **Input**: JSON body
  ```json
  {
    "systolic1": int,
    "diastolic1": int,
    "pulse1": int,
    "systolic2": int,
    "diastolic2": int,
    "pulse2": int,
    "systolic3": int,
    "diastolic3": int,
    "pulse3": int
  }
  ```
- **Returns**: JSON response
  ```json
  {
    "message": "Reading saved successfully",
    "stats": {
      "last_reading": {...},
      "seven_day_avg": {...},
      "thirty_day_avg": {...},
      "all_time_avg": {...}
    },
    "classification": {
      "name": string,
      "description": string,
      "risk": string
    },
    "recommendation": string
  }
  ```
- **Error Cases**:
  - Invalid JSON format (400)
  - Validation errors (400)
  - Database errors (500)
- **Timeout**: 5 seconds for database operations

### 3. Export CSV (`GET /export/csv`)
```go
func (h *Handler) ExportCSVHandler(w http.ResponseWriter, r *http.Request)
```
- **Purpose**: Exports all readings as CSV file
- **Method**: GET
- **Input**: None
- **Returns**: CSV file with headers:
  - Date
  - Time
  - Systolic
  - Diastolic
  - Pulse
  - Classification
- **Headers Set**:
  - Content-Type: text/csv
  - Content-Disposition: attachment
- **Error Cases**:
  - Database errors (500)
  - Writing errors (500)
- **Timeout**: 30 seconds for large datasets

### 4. Static Files (`GET /static/*`)
```go
http.StripPrefix("/static/", http.FileServer(http.Dir("web/static")))
```
- **Purpose**: Serves static assets
- **Method**: GET
- **Serves**:
  - CSS files
  - JavaScript files
- **Path**: web/static directory
- **Error Cases**:
  - File not found (404)

## Security Headers (Applied to all routes)
```go
func (h *Handler) securityHeaders(next http.HandlerFunc) http.HandlerFunc
```
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- X-XSS-Protection: 1; mode=block


## Endpoints

### 1. Home Handler (`GET /`)
```go
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request)
```
- Displays main interface
- Shows current statistics
- Context-aware database queries
- Template rendering

### 2. Submit Reading Handler (`POST /submit`)
```go
func (h *Handler) SubmitReadingHandler(w http.ResponseWriter, r *http.Request)
```
- Accepts JSON input
- Validates readings
- Calculates averages
- Returns updated statistics
- Includes BP classification
- Uses context for timeouts

### 3. Export CSV Handler (`GET /export/csv`)
```go
func (h *Handler) ExportCSVHandler(w http.ResponseWriter, r *http.Request)
```
- Generates CSV download
- Includes all readings
- Proper headers
- Streaming response

## Security Features

### Middleware
```go
func (h *Handler) securityHeaders(next http.HandlerFunc) http.HandlerFunc
```
- X-Content-Type-Options
- X-Frame-Options
- X-XSS-Protection
- Applied to all routes

### Error Handling
```go
func respondWithError(w http.ResponseWriter, message string, code int)
func respondWithJSON(w http.ResponseWriter, data interface{})
```
- Consistent error format
- Safe error messages
- Proper status codes
- JSON responses

## Context Usage
- Request cancellation
- Timeouts
- Database operations
- Template rendering

## Best Practices

### 1. Request Processing
- Validate all inputs
- Use appropriate HTTP methods
- Handle all error cases
- Set correct headers

### 2. Response Handling
- Consistent JSON format
- Proper status codes
- Content-Type headers
- Error wrapping

### 3. Database Operations
- Context-aware queries
- Timeout handling
- Connection management
- Error propagation

### 4. Security
- Input sanitization
- Method validation
- Safe error messages
- Security headers

## Testing Guidelines

### Unit Tests
1. Test all HTTP methods
2. Test input validation
3. Test error scenarios
4. Test context cancellation

### Integration Tests
1. Test database operations
2. Test template rendering
3. Test CSV generation
4. Test security headers

## Usage Example
```go
// Initialize handler
db, _ := database.New("bp.db")
handler, _ := handlers.New(db)

// Get router with all routes configured
mux := handler.Routes()

// Create server
server := &http.Server{
    Addr:    ":32401",
    Handler: mux,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}

// Start server
log.Fatal(server.ListenAndServe())
```

## Error Handling Examples
```go
// Input validation error
if err := validation.ValidateReadings(&input); err != nil {
    respondWithError(w, err.Error(), http.StatusBadRequest)
    return
}

// Database error
if err := h.db.SaveReading(timeoutCtx, avg); err != nil {
    respondWithError(w, "Error saving reading", http.StatusInternalServerError)
    return
}
```

## Future Enhancements
1. Add request logging middleware
2. Implement rate limiting
3. Add metrics collection
4. Expand CSV export options
5. Add data visualization endpoints

## Debugging Tips
1. Check context cancellation
2. Verify timeouts
3. Monitor template execution
4. Watch for database errors
5. Validate security headers
