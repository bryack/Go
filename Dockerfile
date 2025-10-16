# ===== BUILDER STAGE =====
FROM golang:1.24-alpine AS builder
WORKDIR /build
RUN apk add --no-cache gcc musl-dev sqlite-dev
ENV CGO_ENABLED=1
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server
RUN go build -o cli ./cmd/cli

# ===== RUNTIME STAGE =====
FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite-libs
RUN mkdir -p /app
WORKDIR /app
COPY --from=builder /build/server /app/server
COPY --from=builder /build/cli /app/cli
EXPOSE 8080
CMD ["/app/server"]