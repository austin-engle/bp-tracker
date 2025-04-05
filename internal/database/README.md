# Database Package

## Overview
The database package handles all database operations using SQLite.

## Key Files
- `schema.sql`: Database schema definition
- `db.go`: Database interface and operations

## Database Concepts

### Schema Design
```sql
CREATE TABLE IF NOT EXISTS readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    ...
);
```

1. **Primary Key**:
   - `AUTOINCREMENT` handles ID generation
   - Ensures each reading has a unique identifier

2. **Constraints**:
   ```sql
   CONSTRAINT valid_systolic CHECK (systolic BETWEEN 60 AND 250)
   ```
   - Enforces data validity at database level
   - Prevents invalid readings from being saved

3. **Indexes**:
   ```sql
   CREATE INDEX IF NOT EXISTS idx_readings_timestamp ON readings(timestamp);
   ```
   - Improves query performance for time-based lookups
   - Essential for efficient statistics calculations

## Go Database Concepts

1. **Database Connection**:
   ```go
   db, err := sql.Open("sqlite3", dbPath)
   ```
   - Uses `database/sql` package
   - Lazy connection (doesn't test until `Ping()`)
   - Always check errors

2. **Prepared Statements**:
   ```go
   query := `INSERT INTO readings (...) VALUES (?, ?)`
   ```
   - Prevents SQL injection
   - Better performance for repeated queries
   - `?` placeholders are SQLite-specific

3. **Error Handling**:
   ```go
   if err != nil {
       return fmt.Errorf("error saving reading: %w", err)
   }
   ```
   - Uses error wrapping with `%w`
   - Maintains error chain for debugging
   - Provides context at each level

4. **File Embedding**:
   ```go
   //go:embed schema.sql
   var schemaFS embed.FS
   ```
   - Embeds schema file into binary
   - No external files needed
   - New feature in Go 1.16+

## Best Practices
1. Always use prepared statements
2. Check for and handle all errors
3. Close database connections properly
4. Use transactions for multiple operations
5. Add proper indexes for performance
6. Include constraints for data integrity

## Common SQLite Commands
```sql
-- View all readings
SELECT * FROM readings;

-- Get recent readings
SELECT * FROM readings ORDER BY timestamp DESC LIMIT 5;

-- Get averages
SELECT AVG(systolic), AVG(diastolic) FROM readings;
```
