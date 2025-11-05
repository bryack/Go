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
# JWT_SECRET_KEY: Secret key for JWT token signing and verification
#   ⚠️  SECURITY: MUST be overridden when running the container!
#   Usage: docker run -e JWT_SECRET_KEY=your-secure-secret-here ...
ENV JWT_SECRET_KEY="CHANGE_ME_IN_PRODUCTION"

# TASK_DB_PATH: Path to SQLite database file inside the container
#   Default: /data/tasks.db (persisted via volume mount)
#   Override if you need a different database location
ENV TASK_DB_PATH="/data/tasks.db"

# PORT: HTTP server listening port
#   Note: Currently hardcoded in Go code (main.go), this ENV is for future use
#   Default: 8080
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

# Default command: start the HTTP server
CMD ["/app/server"]