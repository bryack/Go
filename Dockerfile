# ===== BUILDER STAGE =====
FROM golang:1.24-alpine AS builder
WORKDIR /build

# Build argument for binary to build
ARG bin_to_build

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev build-base

# Enable CGO for SQLite support
ENV CGO_ENABLED=1

# Download Go module dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build binary
COPY . .
RUN go build -o server ./cmd/${bin_to_build}

# ===== RUNTIME STAGE =====
FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app
COPY --from=builder /build/server /app/server

EXPOSE 8080
CMD ["/app/server"]
