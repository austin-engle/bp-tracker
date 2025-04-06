# File: Dockerfile

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies (removed sqlite-dev)
RUN apk add --no-cache gcc musl-dev

# Copy go module files first for caching
COPY go.mod go.sum ./
RUN go mod download
# Clean module cache to try and resolve build issues
RUN go clean -modcache

# Copy the rest of the source code
COPY . .

# Build the Go binary for Linux AMD64 architecture
# CGO_ENABLED=0 is preferred for Lambda unless CGo is strictly necessary
# Pass -tags lambda.norpc to build statically for AWS Lambda base image
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -tags lambda.norpc -o bootstrap ./cmd/server
# Rename binary to 'bootstrap' as expected by AWS Lambda Go base image

# Final stage - Use AWS Lambda Go base image
FROM public.ecr.aws/lambda/go:1

# WORKDIR is already /var/task (${LAMBDA_TASK_ROOT})

# Copy only the necessary files from builder, specifying absolute destination path
COPY --from=builder /app/bootstrap ${LAMBDA_TASK_ROOT}/
COPY --from=builder /app/web ${LAMBDA_TASK_ROOT}/web/
COPY --from=builder /app/internal/database/schema.sql ${LAMBDA_TASK_ROOT}/schema.sql

# Removed VOLUME instruction
# VOLUME ["/app/data"]

# Removed PORT and DB_PATH environment variables
# ENV PORT=32401 \
#     DB_PATH=/app/data/bp.db

# Removed EXPOSE instruction
# EXPOSE 32401

# Set the CMD to the handler (the compiled binary named bootstrap)
# The base image's entrypoint will execute this.
CMD ["bootstrap"]
