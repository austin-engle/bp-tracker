# Models Package

## Overview
The models package defines the data structures used throughout the application.

## Key Files
- `reading.go`: Defines structures for blood pressure readings

## Data Structures

### Reading
```go
type Reading struct {
    ID            int64     `json:"id"`
    Timestamp     time.Time `json:"timestamp"`
    Systolic      int       `json:"systolic"`
    Diastolic     int       `json:"diastolic"`
    Pulse         int       `json:"pulse"`
    Classification string   `json:"classification"`
}
```

### ReadingInput
```go
type ReadingInput struct {
    Systolic1  int `json:"systolic1"`
    Diastolic1 int `json:"diastolic1"`
    // ... (other fields)
}
```

## Go Concepts Demonstrated

1. **Struct Tags**:
   ```go
   `json:"fieldname"`
   ```
   - Used for JSON serialization/deserialization
   - Tells the JSON encoder what name to use in JSON
   - Common in Go web applications

2. **Time Handling**:
   - Uses Go's built-in `time.Time` type
   - Automatically handles timezone information
   - SQLite compatibility considered

3. **Methods on Types**:
   ```go
   func (ri *ReadingInput) Average() *Reading
   ```
   - Demonstrates Go's method syntax
   - Uses pointer receiver for efficiency
   - Returns a new type (common pattern)

## Best Practices
1. Keep models simple and focused
2. Use appropriate types for each field
3. Include validation methods where appropriate
4. Use meaningful names for fields and methods
