#!/bin/bash

# Get timezone from environment or use default
TZ=${BP_TIMEZONE:-America/Denver}

# Run cleanup script with all arguments and timezone
go run scripts/cleanup.go -tz "$TZ" "$@"
