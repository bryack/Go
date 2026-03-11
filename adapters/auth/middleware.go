package auth

import (
	"context"
	"fmt"
	"log/slog"
	"myproject/internal/handlers"
	"myproject/logger"
	"net/http"
	"strings"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const UserIDKey ContextKey = "user_id"

// AuthMiddleware handles JWT token validation and user authentication for HTTP requests.
type AuthMiddleware struct {
	jwtService *JWTService
	logger     *slog.Logger
}

// NewAuthMiddleware creates a new authentication middleware with the provided JWT service.
func NewAuthMiddleware(jwtService *JWTService, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		logger:     logger,
	}
}

// ExtractToken retrieves and validates the JWT token from the Authorization header.
func (am *AuthMiddleware) ExtractToken(r *http.Request) (token string, err error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	token = parts[1]
	return token, nil
}

// Authenticate wraps an HTTP handler with JWT authentication, adding user ID to request context.
func (am *AuthMiddleware) Authenticate(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := am.ExtractToken(r)
		if err != nil {
			am.logger.Warn("Failed to retrieve or validate token from authorization header",
				slog.String(logger.FieldOperation, "authenticate"),
				slog.String(logger.FieldMethod, r.Method),
				slog.String(logger.FieldPath, r.URL.Path),
				slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
				slog.String(logger.FieldError, err.Error()),
			)
			handlers.JSONError(w, http.StatusUnauthorized, "authorization header required")
			return
		}

		claims, err := am.jwtService.ValidateToken(token)
		if err != nil {
			am.logger.Warn("Failed to validate token",
				slog.String(logger.FieldOperation, "authenticate"),
				slog.String(logger.FieldMethod, r.Method),
				slog.String(logger.FieldPath, r.URL.Path),
				slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
				slog.String(logger.FieldError, err.Error()),
			)
			handlers.JSONError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		userID := claims.UserID
		am.logger.Debug("Authentication successful",
			slog.String(logger.FieldOperation, "authenticate"),
			slog.String(logger.FieldMethod, r.Method),
			slog.String(logger.FieldPath, r.URL.Path),
			slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
			slog.Int(logger.FieldUserID, userID),
		)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		r = r.WithContext(ctx)
		handler(w, r)
	}
}

// GetUserIDFromContext retrieves the authenticated user ID from the request context.
func GetUserIDFromContext(ctx context.Context) (userID int, err error) {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}
