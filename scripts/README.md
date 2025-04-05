# Scripts

## Data Seeding Scripts

### Shell-based Seeding

#### Overview
The `seed.sh` script generates realistic blood pressure readings for testing purposes using the application's HTTP API.

#### Usage
```bash
# Generate test data (default 60 days)
./seed.sh

# Generate specific number of days
./seed.sh --days 30
```

#### Data Ranges
- Systolic: 110-150 mmHg
- Diastolic: 70-90 mmHg
- Pulse: 60-90 bpm
- 3 readings per day
- Generated via HTTP API

#### Options
- `--days`: Number of days to generate (default: 60)

### Go-based Seeding (Legacy)

#### Overview
The Go-based seeding script (`seed.go`) is available for direct database seeding when needed.

#### Usage
```bash
# Basic usage (60 days of data)
go run seed.go

# Custom database path
go run seed.go -db=./mydata.db

# Generate specific number of days
go run seed.go -days=30
```

#### Data Ranges
- Systolic: 110-150 mmHg
- Diastolic: 70-90 mmHg
- Pulse: 60-90 bpm
- 1-3 readings per day
- Readings spaced 4 hours apart

## Data Cleanup Script

### Overview
The cleanup script provides various options for removing blood pressure readings from the database.

### Usage
```bash
# Delete all readings
go run cleanup.go -mode=all

# Delete readings before a specific date
go run cleanup.go -mode=before-date -date=2024-03-01

# Delete readings after a specific date
go run cleanup.go -mode=after-date -date=2024-03-01

# Use with custom database path
go run cleanup.go -db=./mydata.db -mode=all
```

### Options
- `-db`: Database path (default: "bp.db")
- `-mode`: Cleanup mode (default: "all")
  - `all`: Delete all readings
  - `before-date`: Delete readings before specified date
  - `after-date`: Delete readings after specified date
- `-date`: Target date in YYYY-MM-DD format (required for before-date/after-date modes)

### Examples
```bash
# Clean all data
go run cleanup.go

# Remove old readings
go run cleanup.go -mode=before-date -date=2024-01-01

# Remove test data
go run cleanup.go -mode=after-date -date=2024-03-15
```

### Common Use Cases
1. **Reset Database**
   ```bash
   go run scripts/cleanup.go -mode=all
   ```

2. **Remove Old Data**
   ```bash
   go run scripts/cleanup.go -mode=before-date -date=2024-01-01
   ```

3. **Clean and Reseed**
   ```bash
   go run scripts/cleanup.go -mode=all
   ./seed.sh
   ```
