package logutils

import (
	"net/http"

	"go.uber.org/zap"
)

// RequestWithZap returns a shallow copy of parent request with context
// being supplemented with corresponding zap logger.
func RequestWithZap(parent *http.Request, logger *zap.Logger) *http.Request {
	return parent.WithContext(
		ContextWithZap(parent.Context(), logger),
	)
}

// ZapFromRequest retrieves the zap logger passed with request's context.
func ZapFromRequest(request *http.Request) *zap.Logger {
	return ZapFromContext(request.Context())
}
