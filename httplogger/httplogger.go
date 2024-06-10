// Package httplogger implements a middleware that logs the incoming HTTP request & its duration using go-ethereum-log or logrus
package httplogger

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/flashbots/go-utils/logutils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// LoggingMiddleware logs the incoming HTTP request & its duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					method := ""
					url := ""
					if r != nil {
						method = r.Method
						url = r.URL.EscapedPath()
					}

					log.Error(fmt.Sprintf("http request panic: %s %s", method, url),
						"err", err,
						"trace", string(debug.Stack()),
					)
				}
			}()
			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			log.Info(fmt.Sprintf("http: %s %s %d", r.Method, r.URL.EscapedPath(), wrapped.status),
				"status", wrapped.status,
				"method", r.Method,
				"path", r.URL.EscapedPath(),
				"duration", fmt.Sprintf("%f", time.Since(start).Seconds()),
			)
		},
	)
}

// LoggingMiddlewareSlog logs the incoming HTTP request & its duration.
func LoggingMiddlewareSlog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					method := ""
					url := ""
					if r != nil {
						method = r.Method
						url = r.URL.EscapedPath()
					}

					logger.Error(fmt.Sprintf("http request panic: %s %s", method, url),
						"err", err,
						"trace", string(debug.Stack()),
						"method", r.Method,
					)
				}
			}()
			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			logger.Info(fmt.Sprintf("http: %s %s %d", r.Method, r.URL.EscapedPath(), wrapped.status),
				"status", wrapped.status,
				"method", r.Method,
				"path", r.URL.EscapedPath(),
				"duration", fmt.Sprintf("%f", time.Since(start).Seconds()),
			)
		},
	)
}

// LoggingMiddlewareLogrus logs the incoming HTTP request & its duration.
func LoggingMiddlewareLogrus(logger *logrus.Entry, next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					method := ""
					url := ""
					if r != nil {
						method = r.Method
						url = r.URL.EscapedPath()
					}

					logger.WithFields(logrus.Fields{
						"err":    err,
						"trace":  string(debug.Stack()),
						"method": r.Method,
					}).Error(fmt.Sprintf("http request panic: %s %s", method, url))
				}
			}()
			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			logger.WithFields(logrus.Fields{
				"status":   wrapped.status,
				"method":   r.Method,
				"path":     r.URL.EscapedPath(),
				"duration": fmt.Sprintf("%f", time.Since(start).Seconds()),
			}).Info(fmt.Sprintf("http: %s %s %d", r.Method, r.URL.EscapedPath(), wrapped.status))
		},
	)
}

// LoggingMiddlewareZap logs the incoming HTTP request & its duration.
func LoggingMiddlewareZap(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID (`base64` to shorten its string representation)
		_uuid := [16]byte(uuid.New())
		httpRequestID := base64.RawStdEncoding.EncodeToString(_uuid[:])

		l := logger.With(
			zap.String("httpRequestID", httpRequestID),
			zap.String("logType", "activity"),
		)
		r = logutils.RequestWithZap(r, l)

		// Handle panics
		defer func() {
			if msg := recover(); msg != nil {
				w.WriteHeader(http.StatusInternalServerError)
				var method, url string
				if r != nil {
					method = r.Method
					url = r.URL.EscapedPath()
				}
				l.Error("HTTP request handler panicked",
					zap.Any("error", msg),
					zap.String("method", method),
					zap.String("url", url),
				)
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(w, r)

		// Passing request stats both in-message (for the human reader)
		// as well as inside the structured log (for the machine parser)
		logger.Info(fmt.Sprintf("%s: %s %s %d", r.URL.Scheme, r.Method, r.URL.EscapedPath(), wrapped.status),
			zap.Int("durationMs", int(time.Since(start).Milliseconds())),
			zap.Int("status", wrapped.status),
			zap.String("httpRequestID", httpRequestID),
			zap.String("logType", "access"),
			zap.String("method", r.Method),
			zap.String("path", r.URL.EscapedPath()),
		)
	})
}
