package internal

import (
	"context"
	"log/slog"
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
