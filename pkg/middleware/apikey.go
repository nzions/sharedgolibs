// SPDX-License-Identifier: CC0-1.0

package middleware

import "net/http"

// WithAPIKey adds API key authentication to HTTP handlers.
// Checks for the API key in the X-API-Key header or api_key query parameter.
// Returns HTTP 401 Unauthorized if the key does not match.
func WithAPIKey(key string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip API key check if no key is configured
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for API key in header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Check for API key in query parameter
			apiKey = r.URL.Query().Get("api_key")
		}

		if apiKey != key {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized: Invalid or missing API key"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
