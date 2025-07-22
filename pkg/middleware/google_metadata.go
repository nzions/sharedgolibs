// SPDX-License-Identifier: CC0-1.0

package middleware

import "net/http"

// WithGoogleMetadataFlavor adds Google Cloud metadata flavor headers to HTTP responses.
// Adds standard headers that identify responses as coming from Google Cloud metadata service.
// Example usage:
//
//	handler := middleware.WithGoogleMetadataFlavor(myHandler)
func WithGoogleMetadataFlavor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Metadata-Flavor", "Google")
		w.Header().Set("Server", "Metadata Server for Google Compute Engine")

		next.ServeHTTP(w, r)
	})
}
