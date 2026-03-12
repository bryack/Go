package webserver

import (
	"context"
	"fmt"
	"log/slog"
	"myproject/application"
	"myproject/domain"
	"myproject/logger"
	"net/http"
	"strings"
)

// AuthMiddleware handles JWT token validation and user authentication for HTTP requests.
type AuthMiddleware struct {
	tokenGenerator domain.TokenGenerator
	logger         *slog.Logger
}

// NewAuthMiddleware creates a new authentication middleware with the provided JWT service.
func NewAuthMiddleware(tokenGenerator domain.TokenGenerator, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		tokenGenerator: tokenGenerator,
		logger:         logger,
	}
}

// extractToken retrieves and validates the JWT token from the Authorization header.
func (am *AuthMiddleware) extractToken(r *http.Request) (token string, err error) {
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
		token, err := am.extractToken(r)
		if err != nil {
			am.logger.Warn("Failed to retrieve or validate token from authorization header",
				slog.String(logger.FieldOperation, "authenticate"),
				slog.String(logger.FieldMethod, r.Method),
				slog.String(logger.FieldPath, r.URL.Path),
				slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
				slog.String(logger.FieldError, err.Error()),
			)
			JSONError(w, http.StatusUnauthorized, "authorization header required")
			return
		}

		claims, err := am.tokenGenerator.ValidateToken(token)
		if err != nil {
			am.logger.Warn("Failed to validate token",
				slog.String(logger.FieldOperation, "authenticate"),
				slog.String(logger.FieldMethod, r.Method),
				slog.String(logger.FieldPath, r.URL.Path),
				slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
				slog.String(logger.FieldError, err.Error()),
			)
			JSONError(w, http.StatusUnauthorized, "invalid or expired token")
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

		ctx := context.WithValue(r.Context(), application.UserIDKey, userID)
		r = r.WithContext(ctx)
		handler(w, r)
	}
}
