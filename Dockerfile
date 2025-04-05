# File: Dockerfile

# Build stage
FROM golang:1.22.5-alpine AS builder

WORKDIR /app

# Install sqlite and build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy the entire source code
COPY . .

# Download dependencies and build
RUN go mod download && \
    CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o bp-tracker ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache sqlite-libs tzdata

# Set the timezone to MST
ENV TZ=America/Denver
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy only the necessary files from builder
COPY --from=builder /app/bp-tracker .
COPY --from=builder /app/web ./web

# Create volume for database
VOLUME ["/app/data"]

# Set environment variables
ENV PORT=32401 \
    DB_PATH=/app/data/bp.db

EXPOSE 32401

CMD ["./bp-tracker"]
