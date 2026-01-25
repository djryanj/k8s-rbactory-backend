// compression.go
package middleware

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

const (
	// DefaultMinCompressionSize is the minimum response size (in bytes) that will be compressed
	// Responses smaller than this are sent uncompressed as the overhead isn't worth it
	DefaultMinCompressionSize = 1024 // 1KB
)

// gzipResponseWriter wraps http.ResponseWriter to support gzip compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write(b)
}

// Flush implements http.Flusher
func (w *gzipResponseWriter) Flush() {
	if gw, ok := w.Writer.(*gzip.Writer); ok {
		gw.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// gzipWriterPool reuses gzip writers for better performance
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

// Compression returns a middleware that compresses HTTP responses using gzip
// when the client supports it and the response is compressible.
// Uses a default minimum size of 1KB - responses smaller than this are not compressed.
func Compression(logger *slog.Logger) func(http.Handler) http.Handler {
	return CompressionWithMinSize(logger, DefaultMinCompressionSize)
}

// CompressionWithMinSize returns a middleware that only compresses responses
// larger than the specified minimum size (in bytes)
func CompressionWithMinSize(logger *slog.Logger, minSize int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip encoding
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Create a buffer to capture the response
			crw := &capturingResponseWriter{
				ResponseWriter: w,
				minSize:        minSize,
			}

			next.ServeHTTP(crw, r)

			// If the response is large enough, compress it
			if len(crw.body) >= minSize {
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set("Vary", "Accept-Encoding")
				w.Header().Del("Content-Length")

				gz := gzipWriterPool.Get().(*gzip.Writer)
				defer gzipWriterPool.Put(gz)

				gz.Reset(w)
				defer func() {
					if err := gz.Close(); err != nil {
						logger.Error("Failed to close gzip writer", "error", err)
					}
				}()

				if crw.statusCode != 0 {
					w.WriteHeader(crw.statusCode)
				}

				if _, err := gz.Write(crw.body); err != nil {
					logger.Error("Failed to write compressed response", "error", err)
				}
			} else {
				// Response is too small, send uncompressed
				if crw.statusCode != 0 {
					w.WriteHeader(crw.statusCode)
				}
				w.Write(crw.body)
			}
		})
	}
}

// capturingResponseWriter captures the response for conditional compression
type capturingResponseWriter struct {
	http.ResponseWriter
	body       []byte
	statusCode int
	minSize    int
}

func (w *capturingResponseWriter) WriteHeader(status int) {
	w.statusCode = status
}

func (w *capturingResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}
