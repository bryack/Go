// Package logger provides structured logging utilities with standard
// field names and data masking for sensitive information.
package logger

import (
	"strings"
)

const (
	FieldRequestID  = "request_id"
	FieldUserID     = "user_id"
	FieldMethod     = "method"
	FieldPath       = "path"
	FieldStatusCode = "status_code"
	FieldDuration   = "duration_ms"
	FieldError      = "error"
	FieldOperation  = "operation"
	FieldTaskID     = "task_id"
	FieldEmail      = "email" // Always masked
	FieldTraceID    = "trace_id"
	FieldSpanID     = "span_id"
)

// MaskEmail masks an email address for privacy protection.
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")

	if len(parts) != 2 {
		return "***"
	}

	userName := parts[0]
	domain := parts[1]

	if len(userName) <= 2 {
		return "***" + "@" + domain
	}
	return userName[:1] + "***" + userName[len(userName)-1:] + "@" + domain
}

// MaskToken masks authentication tokens and API keys for security.
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
