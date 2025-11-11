package logger

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// recoverPanic recovers from panics in HTTP handlers, logs the error with stack trace,
// and returns a 500 Internal Server Error response to the client.
func recoverPanic(logger *slog.Logger, w http.ResponseWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		stackTrace := string(debug.Stack())
		requestID := GetRequestID(r.Context())

		logger.Error("Panic recovered",
			slog.Any("panic", rec),                 // The panic value
			slog.String("stack_trace", stackTrace), // Full stack trace
			slog.String(FieldRequestID, requestID), // Request ID
			slog.String(FieldMethod, r.Method),     // HTTP method
			slog.String(FieldPath, r.URL.Path),     // Request path
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// LoggingMiddleware returns HTTP middleware that logs request start/completion with structured fields.
// Generates unique request IDs for correlation and includes method, path, duration, and user_agent in logs.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate request ID and add to context
			requestID := GenerateRequestID()
			ctx := WithRequestID(r.Context(), requestID)
			r = r.WithContext(ctx)

			// Record start time
			start := time.Now()

			// Log request start
			logger.Info("HTTP request started",
				slog.String(FieldRequestID, requestID),
				slog.String(FieldMethod, r.Method),
				slog.String(FieldPath, r.URL.Path),
				slog.String("user_agent", r.UserAgent()),
			)

			// Set up panic recovery
			defer recoverPanic(logger, w, r)

			// Call the next handler
			next.ServeHTTP(w, r)

			// Calculate duration
			duration := time.Since(start).Milliseconds()

			// Log request completion
			logger.Info("HTTP request completed",
				slog.String(FieldRequestID, requestID),
				slog.String(FieldMethod, r.Method),
				slog.String(FieldPath, r.URL.Path),
				slog.Int64(FieldDuration, duration),
			)
		})
	}
}
