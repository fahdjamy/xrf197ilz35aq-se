package server

import (
	"context"
	"log/slog"
	"xrf197ilz35aq/internal/model"
)

// ContextKey is a custom type to avoid key collisions in the context map.
type ContextKey string

const LoggerContextKey = ContextKey("logger")

// LoggerFromContext is a helper function to retrieve the logger from the context.
// It ensures type safety and returns a default logger if none is found.
func LoggerFromContext(ctx context.Context, defaultLogger slog.Logger) *slog.Logger {
	// ctx.Value returns an interface{}, so we need to do a type assertion.
	if logger, ok := ctx.Value(LoggerContextKey).(*slog.Logger); ok {
		return logger
	}
	// Return the default logger if no logger is found in the context.
	return &defaultLogger
}

func CreateAuthTokenHeader(token string) map[string]string {
	xrfAuthToken := "xrf-auth-token"
	return map[string]string{
		xrfAuthToken: token,
	}
}

const UserContextKey = ContextKey("user-context")

// ContextWithUserCtx returns a new context with the given user object.
func ContextWithUserCtx(ctx context.Context, user *model.UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// UserFromContext retrieves the user object from the context.
// It returns the user and a boolean indicating if the user was found.
func UserFromContext(ctx context.Context) (*model.UserContext, bool) {
	user, ok := ctx.Value(UserContextKey).(*model.UserContext)
	return user, ok
}
