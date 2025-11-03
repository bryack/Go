package auth

import (
	"context"
	"fmt"
	"myproject/internal/handlers"
	"net/http"
	"strings"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

type AuthMiddleware struct {
	jwtService *JWTService
}

func NewAuthMiddleware(jwtService *JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

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

func (am *AuthMiddleware) Authenticate(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := am.ExtractToken(r)
		if err != nil {
			handlers.JSONError(w, http.StatusUnauthorized, "authorization header required")
			return
		}

		claims, err := am.jwtService.ValidateToken(token)
		if err != nil {
			handlers.JSONError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		userID := claims.UserID
		ctx := context.WithValue(context.Background(), UserIDKey, userID)

		handler(w, r)
	}
}
