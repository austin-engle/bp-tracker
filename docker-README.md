# Docker Configuration

## Overview
The application uses a multi-stage Docker build to create a minimal production image.

## Quick Start
```bash
# Build and start the application
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the application
docker-compose down
```

## Key Features
- Multi-stage build for smaller image size
- Persistent volume for database storage
- Alpine-based for security and size
- Proper handling of SQLite dependencies

## Data Persistence
Database is stored in a Docker volume `bp-data`.
To backup your data:
```bash
docker run --rm -v bp-tracker_bp-data:/data -v $(pwd):/backup alpine tar czf /backup/bp-backup.tar.gz /data
```
