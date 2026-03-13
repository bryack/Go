package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"myproject/application"
	"myproject/domain"
	"myproject/logger"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor handles JWT token validation and user authentication for gRPC requests
type AuthInterceptor struct {
	tokenGenerator domain.TokenGenerator
	logger         *slog.Logger
}

// NewAuthInterceptor creates a new authentication interceptor with the provided JWT service
func NewAuthInterceptor(tokenGenerator domain.TokenGenerator, logger *slog.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		tokenGenerator: tokenGenerator,
		logger:         logger,
	}
}

func (a *AuthInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if a.isPublicMethod(info.FullMethod) {
		return handler(ctx, req)
	}

	token, err := a.extractToken(ctx)
	if err != nil {
		a.logger.Warn("Failed to retrieve token from authorization header",
			slog.String(logger.FieldOperation, "auth_interceptor"),
			slog.String(logger.FieldMethod, info.FullMethod),
			slog.String(logger.FieldError, err.Error()),
		)
		return nil, status.Error(codes.Unauthenticated, "authorization header required")
	}

	claims, err := a.tokenGenerator.ValidateToken(token)
	if err != nil {
		a.logger.Warn("Failed to validate token",
			slog.String(logger.FieldOperation, "auth_interceptor"),
			slog.String(logger.FieldMethod, info.FullMethod),
			slog.String(logger.FieldError, err.Error()),
		)
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	userID := claims.UserID
	a.logger.Debug("Authentication successful",
		slog.String(logger.FieldOperation, "auth_interceptor"),
		slog.String(logger.FieldMethod, info.FullMethod),
		slog.Int(logger.FieldUserID, userID),
	)

	ctx = context.WithValue(ctx, application.UserIDKey, userID)

	return handler(ctx, req)
}

func (a *AuthInterceptor) extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return "", fmt.Errorf("authorization header required")
	}

	parts := strings.SplitN(authHeader[0], " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return parts[1], nil
}

func (a *AuthInterceptor) isPublicMethod(fullMethod string) bool {
	switch fullMethod {
	case "/grpcserver.TaskManager/Register",
		"/grpcserver.TaskManager/Login":
		return true
	}
	return false
}
