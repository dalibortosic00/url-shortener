package request

import "context"

type contextKey string

const userContextKey = contextKey("user")

// UserID retrieves the user ID from the context, if available.
func UserID(ctx context.Context) string {
	id, _ := ctx.Value(userContextKey).(string)
	return id
}

// WithUserID returns a new context with the user ID set.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userContextKey, id)
}
