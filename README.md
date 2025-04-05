# Blood Pressure Tracker

## Project Overview
This application helps track blood pressure readings over time, providing averages and health classifications based on the American Heart Association guidelines.

## Technical Requirements
- Go 1.22.5 or later
- SQLite 3
- Docker (optional)

## Key Features
- Modern Go 1.22 routing
- Context-aware operations
- Security middleware
- Graceful shutdown
- CSV export capability

## Project Structure
```
bp-tracker/
├── cmd/            # Application entry points
├── internal/       # Private application code
│   ├── database/  # SQLite operations
│   ├── handlers/  # HTTP handlers (Go 1.22 patterns)
│   ├── models/    # Data structures
│   ├── utils/     # BP classification
│   └── validation/# Input validation
├── web/           # Frontend assets
└── scripts/       # Utility scripts
```

## Key Concepts
1. **Go Modules**: The project uses Go modules for dependency management.
   - `go.mod` defines our module and dependencies
   - Run `go mod tidy` to manage dependencies

2. **Project Organization**:
   - `cmd/`: Contains main applications
   - `internal/`: Private application code (cannot be imported by other projects)
   - `web/`: Web-related assets
   - Each component has its own README with detailed explanations

3. **Privacy in Go**:
   - Uppercase names are exported (public)
   - Lowercase names are unexported (private)
   - Code in `internal/` cannot be imported by other projects

## Getting Started
1. Install Go 1.22.5+
2. Clone repository
3. Run:
   ```bash
   go mod tidy
   go run cmd/server/main.go
   ```
   Or with Docker:
   ```bash
   docker-compose up -d
   ```

## Development
- Each directory contains a README explaining its purpose
- Code is documented with comments explaining key concepts
- Follow Go best practices (gofmt, golint)
- Uses Go 1.22 features
- Modern error handling
- Context-aware operations
- Proper shutdown handling

### Test Data
You can generate test data using the seed script:
```bash
# Generate test data (default 60 days)
./scripts/seed.sh

# Generate specific number of days
./scripts/seed.sh --days 30
docker-compose -f docker-compose.dev.yml exec bp-tracker /app/scripts/seed.sh --days 30```

## Security
- Security headers enabled
- Input validation
- Safe SQL operations
