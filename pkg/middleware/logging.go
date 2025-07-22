// SPDX-License-Identifier: CC0-1.0

package middleware

import (
	"log"
	"log/slog"
	"net/http"
	"time"
)

// ResponseRecorder wraps http.ResponseWriter to capture status code.
type ResponseRecorder struct {
	http.ResponseWriter
	Status int
}

func (rw *ResponseRecorder) WriteHeader(code int) {
	rw.Status = code
	rw.ResponseWriter.WriteHeader(code)
}

// WithLogging creates a logging middleware that works with both log.Logger and slog.Logger.
// Returns a middleware handler that logs requests and responses.
func WithLogging(logger interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &ResponseRecorder{
				ResponseWriter: w,
				Status:         http.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			duration := time.Since(start)

			// Log using the appropriate logger type
			switch l := logger.(type) {
			case *log.Logger:
				l.Printf("%s %s %s %d %s", r.Method, r.URL.Path, r.RemoteAddr, recorder.Status, duration)
			case *slog.Logger:
				l.Info("HTTP request",
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
					"status", recorder.Status,
					"duration", duration,
				)
			default:
				// Fallback to standard log
				log.Printf("%s %s %s %d %s", r.Method, r.URL.Path, r.RemoteAddr, recorder.Status, duration)
			}
		})
	}
}
