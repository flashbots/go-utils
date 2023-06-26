// Package logutils implements helpers for logging.
package logutils

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const loggerContextKey contextKey = "logger"

// ContextWithZap returns a copy of parent context injected with corresponding
// zap logger.
func ContextWithZap(parent context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(parent, loggerContextKey, logger)
}

// ZapFromContext retrieves the zap logger passed with a context.
func ZapFromContext(ctx context.Context) *zap.Logger {
	if l, found := ctx.Value(loggerContextKey).(*zap.Logger); found {
		return l
	}
	return zap.L()
}
