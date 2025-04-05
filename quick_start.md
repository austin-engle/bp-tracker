# Quick Start Guide: Blood Pressure Tracker

## Development Setup

### 1. Prerequisites
```bash
# Required
- Go 1.22.5 or later
- SQLite 3
- Docker & Docker Compose (optional)

# Check Go version
go version
```

### 2. Clone and Setup
```bash
# Create project directory
mkdir bp-tracker
cd bp-tracker

# Copy all provided files into their respective directories
# Ensure directory structure matches the project layout
```

### 3. Local Development
```bash
# Install dependencies
go mod tidy

# Run the application
go run cmd/server/main.go

# Access the application
open http://localhost:32401
```

### 4. Docker Development
```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the application
docker-compose down
```

### 5. Test Data Setup
```bash
# Generate test data (60 days)
go run scripts/seed.go

# For Docker:
docker-compose exec bp-tracker go run scripts/seed.go
```

### 6. Common Operations
```bash
# Clean all data
go run scripts/cleanup.go -mode=all

# Export data
curl http://localhost:32401/export/csv > readings.csv

# Monitor logs
tail -f bp-tracker.log
```

## Testing the Setup

1. Initial Test
```bash
# Start the application
go run cmd/server/main.go

# In another terminal, test the API
curl http://localhost:32401
```

2. Add Test Reading
```bash
curl -X POST http://localhost:32401/submit \
  -H "Content-Type: application/json" \
  -d '{
    "systolic1": 120,
    "diastolic1": 80,
    "pulse1": 72,
    "systolic2": 122,
    "diastolic2": 81,
    "pulse2": 70,
    "systolic3": 121,
    "diastolic3": 80,
    "pulse3": 71
  }'
```

## Troubleshooting

### Common Issues
1. Port already in use:
```bash
# Check port usage
lsof -i :32401
# Use different port
go run cmd/server/main.go -port=32402
```

2. Database issues:
```bash
# Reset database
go run scripts/cleanup.go -mode=all
# Check permissions
ls -l bp.db
```

3. Docker issues:
```bash
# Remove containers and volumes
docker-compose down -v
# Rebuild
docker-compose up -d --build
```

### Health Check
```bash
# Check application status
curl http://localhost:32401/

# Check database
sqlite3 bp.db "SELECT count(*) FROM readings;"
```
