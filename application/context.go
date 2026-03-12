package application

import (
	"context"
	"fmt"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const UserIDKey ContextKey = "user_id"

// GetUserIDFromContext retrieves the authenticated user ID from the request context.
func GetUserIDFromContext(ctx context.Context) (userID int, err error) {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}
