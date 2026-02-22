# ===== BUILDER STAGE =====
FROM golang:1.24-alpine AS builder
WORKDIR /build
RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/grpcserver

# ===== RUNTIME STAGE =====
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /build/server /app/server

EXPOSE 50051
CMD ["/app/server"]