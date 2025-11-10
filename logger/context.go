// Package logger provides context helpers for request correlation.
// Functions for storing/retrieving request IDs, trace IDs, and loggers
// enable log correlation across the entire request lifecycle.
package logger

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
