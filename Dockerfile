# ===== BUILDER STAGE =====
# Multi-stage build: compile Go binaries with all build dependencies
FROM golang:1.24-alpine AS builder
WORKDIR /build

# Install build dependencies required for CGO and SQLite compilation
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Enable CGO for SQLite support (modernc.org/sqlite requires CGO)
ENV CGO_ENABLED=1

# Download Go module dependencies (cached layer for faster rebuilds)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build server binary
COPY . .
RUN go build -o server ./cmd/server
# TODO: CLI build disabled - needs refactoring for new storage interface (userID parameter)
# RUN go build -o cli ./cmd/cli

# ===== RUNTIME STAGE =====
# Minimal runtime image with only necessary dependencies
FROM alpine:latest

# Install runtime dependencies: ca-certificates for HTTPS, sqlite-libs for database
RUN apk add --no-cache ca-certificates sqlite-libs

# Create application and data directories
RUN mkdir -p /app
RUN mkdir -p /data
WORKDIR /app

# Environment Variables Configuration
# ---------------------------------
# The server supports multiple configuration methods with the following precedence:
#   1. Command-line flags (highest priority)
#   2. Environment variables with TASKMANAGER_ prefix
#   3. Configuration file (config.yaml)
#   4. Default values (lowest priority)
#
# NEW CONFIGURATION (Recommended):
# Use TASKMANAGER_ prefixed environment variables for structured configuration:
#
# TASKMANAGER_JWT_SECRET: Secret key for JWT token signing and verification
#   ⚠️  SECURITY: MUST be overridden when running the container!
#   Must be at least 32 characters long
#   Usage: docker run -e TASKMANAGER_JWT_SECRET=your-secure-secret-here ...
#
# TASKMANAGER_JWT_EXPIRATION: JWT token expiration duration
#   Default: 24h
#   Format: Go duration string (e.g., "24h", "48h", "30m")
#   Usage: docker run -e TASKMANAGER_JWT_EXPIRATION=48h ...
#
# TASKMANAGER_DATABASE_PATH: Path to SQLite database file inside the container
#   Default: ./data/tasks.db
#   Usage: docker run -e TASKMANAGER_DATABASE_PATH=/data/tasks.db ...
#
# TASKMANAGER_SERVER_PORT: HTTP server listening port
#   Default: 8080
#   Usage: docker run -e TASKMANAGER_SERVER_PORT=8080 ...
#
# TASKMANAGER_SERVER_HOST: HTTP server host address
#   Default: 0.0.0.0
#   Usage: docker run -e TASKMANAGER_SERVER_HOST=0.0.0.0 ...
#
# LEGACY ENVIRONMENT VARIABLES (Deprecated, but still supported):
# These are maintained for backward compatibility and will show deprecation warnings:
#
# JWT_SECRET_KEY: Maps to TASKMANAGER_JWT_SECRET (deprecated)
#   ⚠️  Use TASKMANAGER_JWT_SECRET instead
ENV JWT_SECRET_KEY="CHANGE_ME_IN_PRODUCTION"

# TASK_DB_PATH: Maps to TASKMANAGER_DATABASE_PATH (deprecated)
#   ⚠️  Use TASKMANAGER_DATABASE_PATH instead
ENV TASK_DB_PATH="/data/tasks.db"

# PORT: Maps to TASKMANAGER_SERVER_PORT (deprecated)
#   ⚠️  Use TASKMANAGER_SERVER_PORT instead
ENV PORT=8080

# Copy compiled binary from builder stage
COPY --from=builder /build/server /app/server
# TODO: CLI binary disabled - uncomment when CLI is refactored
# COPY --from=builder /build/cli /app/cli

# Expose HTTP server port
EXPOSE 8080

# Declare /data as a volume for database persistence
# Mount a host directory to persist data across container restarts:
#   docker run -v /host/path:/data ...
VOLUME ["/data"]

# Configuration File Support
# --------------------------
# You can mount a configuration file (config.yaml or config.json) to configure the server:
#
# Example with YAML config file:
#   docker run -v /host/path/config.yaml:/app/config.yaml myapp
#
# Example with JSON config file:
#   docker run -v /host/path/config.json:/app/config.json myapp
#
# The server searches for config files in the following locations:
#   1. Current directory (./config.yaml or ./config.json)
#   2. /etc/taskmanager/
#   3. ~/.taskmanager/
#
# Example config.yaml structure:
#   server:
#     port: 8080
#     host: 0.0.0.0
#   database:
#     path: /data/tasks.db
#   jwt:
#     secret: your-secret-key-here
#     expiration: 24h

# Default command: start the HTTP server
# 
# Command-Line Flag Examples:
# ---------------------------
# You can override any configuration using command-line flags:
#
# Basic usage with flags:
#   docker run myapp --jwt-secret=your-secret --port=8080
#
# Show current configuration:
#   docker run myapp --show-config
#
# Use custom config file:
#   docker run -v /host/config.yaml:/app/config.yaml myapp --config=/app/config.yaml
#
# Override specific settings:
#   docker run myapp --jwt-secret=secret --jwt-expiration=48h --db-path=/data/tasks.db
#
# Available flags:
#   --port              HTTP server port (default: 8080)
#   --host              HTTP server host (default: 0.0.0.0)
#   --db-path           Path to SQLite database file (default: ./data/tasks.db)
#   --jwt-secret        JWT secret key (required, min 32 chars)
#   --jwt-expiration    JWT token expiration duration (default: 24h)
#   --config            Path to configuration file
#   --show-config       Display current configuration and exit
#   --help              Show help message
#
# Complete example combining all methods:
#   docker run \
#     -e TASKMANAGER_JWT_SECRET=my-secret-key \
#     -v /host/data:/data \
#     -v /host/config.yaml:/app/config.yaml \
#     -p 8080:8080 \
#     myapp --port=8080 --show-config
#
CMD ["/app/server"]