// SPDX-License-Identifier: CC0-1.0

package middleware

import (
	"net/http"

	"github.com/nzions/sharedgolibs/pkg/logi"
)

// LogAndCORS combines logging and CORS middleware in a single handler.
// Convenience function for the common pattern of applying both middlewares.
func LogAndCORS(logger logi.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return WithLogging(logger)(WithCORS(next))
	}
}

// Chain applies multiple middleware functions in order.
// Returns a handler with all middleware applied.
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
