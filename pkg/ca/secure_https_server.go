package ca

import (
	"fmt"
	"net/http"
)

// SecureHTTPSServer wraps http.Server to ensure only HTTPS methods are used
// and prevents accidental HTTP ListenAndServe usage.
type SecureHTTPSServer struct {
	*http.Server
}

// ListenAndServe is intentionally not implemented to prevent HTTP usage
// Use ListenAndServeTLS() for HTTPS servers
func (s *SecureHTTPSServer) ListenAndServe() error {
	return fmt.Errorf("cannot call ListenAndServe() on SecureHTTPSServer - use ListenAndServeTLS()")
}

// NewSecureHTTPSServer wraps the CA helper to return a type-safe server
func NewSecureHTTPSServer(server *http.Server) *SecureHTTPSServer {
	return &SecureHTTPSServer{Server: server}
}
