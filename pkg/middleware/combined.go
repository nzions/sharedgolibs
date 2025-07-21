package middleware

import "net/http"

// LogAndCORS combines logging and CORS middleware in a single handler.
// This is a convenience function for the common pattern of applying both middlewares.
func LogAndCORS(logger interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return WithLogging(logger)(WithCORS(next))
	}
}

// Chain applies multiple middleware functions in order
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
