package utils

import "context"

type contextKey string

const UserIDKey contextKey = "userID"

// GetUserIDFromContext is a helper for any handler to easily get the logged in User ID.
func GetUserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}
