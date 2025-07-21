// SPDX-License-Identifier: CC0-1.0

package middleware

import "net/http"

// WithGoogleMetadataFlavor adds Google Cloud metadata flavor headers to HTTP responses.
// This middleware adds the standard headers that identify responses as coming from
// Google Cloud metadata service, useful for applications that expect these headers.
//
// Example:
//
//	handler := middleware.WithGoogleMetadataFlavor(myHandler)
func WithGoogleMetadataFlavor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Metadata-Flavor", "Google")
		w.Header().Set("Server", "Metadata Server for Google Compute Engine")

		next.ServeHTTP(w, r)
	})
}
