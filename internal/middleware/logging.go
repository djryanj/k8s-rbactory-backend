// backend/internal/middleware/logging.go
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Unwrap returns the original ResponseWriter for interface checks
// (e.g., http.Hijacker, http.Flusher, http.Pusher)
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// Logging middleware using structured logging with context
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			// Create request-scoped logger with request metadata
			requestID := logging.GetRequestID(r.Context())
			requestLogger := logging.NewRequestLogger(logger, requestID, r.Method, r.URL.Path)

			// Add logger to context for downstream handlers
			ctx := logging.WithLogger(r.Context(), requestLogger)

			// Recover from panics
			defer func() {
				if err := recover(); err != nil {
					requestLogger.Error("panic recovered",
						slog.Any("error", err),
						slog.String("remote_addr", r.RemoteAddr),
					)
					if !wrapped.wroteHeader {
						http.Error(wrapped, "Internal Server Error", http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(wrapped, r.WithContext(ctx))

			duration := time.Since(start)

			// Choose log level based on status code
			logLevel := slog.LevelInfo
			if wrapped.status >= 500 {
				logLevel = slog.LevelError
			} else if wrapped.status >= 400 {
				logLevel = slog.LevelWarn
			}

			requestLogger.Log(r.Context(), logLevel, "http request completed",
				slog.Int("status", wrapped.status),
				slog.Int("size", wrapped.size),
				slog.Duration("duration", duration),
				slog.Int64("duration_ms", duration.Milliseconds()),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("referer", r.Referer()),
			)
		})
	}
}
