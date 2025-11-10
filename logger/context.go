// Package logger provides context helpers for request correlation.
// Functions for storing/retrieving request IDs, trace IDs, and loggers
// enable log correlation across the entire request lifecycle.
package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"
)

type contextKey int

const (
	requestIDKey contextKey = iota
	traceIDKey
	loggerKey
)

// generateRequestID creates a unique request ID combining timestamp and random data.
func generateRequestID() string {
	timeNow := time.Now().UnixMilli()
	randomBytes := make([]byte, 8)

	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Sprintf("req_error_%d", timeNow)
	}

	return fmt.Sprintf("req_%d_%s", timeNow, hex.EncodeToString(randomBytes))
}

// WithRequestID stores a request ID in the context for request correlation.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}

	return requestID
}

// WithTraceID stores a trace ID in the context for distributed tracing.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID retrieves the trace ID from the context.
func GetTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(traceIDKey).(string)
	if !ok {
		return ""
	}

	return traceID
}

// WithLogger stores a logger instance in the context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves a logger from the context.
func FromContext(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return log
}
