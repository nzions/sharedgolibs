// SPDX-License-Identifier: CC0-1.0

package ca

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nzions/sharedgolibs/pkg/ca/dualprotocol"
	"github.com/nzions/sharedgolibs/pkg/logi"
)

// CreateSecureDualProtocolServer creates a dual protocol server with certificates from the CA.
// This is a convenience method that requests certificates from the CA server and
// returns a configured server ready to serve both HTTP and HTTPS traffic on the same port.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL for certificate requests
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Parameters:
//   - serviceName: Name of the service for certificate generation
//   - port: Port number the server will listen on (without ":")
//   - sans: Subject Alternative Names - mix of domain names and IP addresses
//     The CA automatically detects IPs vs hostnames and places them in the correct certificate fields
//     The first non-IP entry will be used as the Common Name
//   - handler: HTTP handler for the server (if nil, uses default handler)
//   - logger: Logger instance (if nil, creates a new daemon logger)
//
// Returns a configured server with TLS certificates, ready to call ListenAndServe().
func CreateSecureDualProtocolServer(serviceName, port string, sans []string, handler http.Handler, logger logi.Logger) (*dualprotocol.Server, error) {
	// Request certificate from CA using simplified V2 API with automatic IP detection and CN selection
	certResp, err := RequestCertificateV2(serviceName, sans)
	if err != nil {
		return nil, fmt.Errorf("failed to request certificate: %w", err)
	}

	// Parse the certificate and key
	cert, err := tls.X509KeyPair([]byte(certResp.Certificate), []byte(certResp.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   serviceName + ".local",
	}

	// Use default logger if none provided
	if logger == nil {
		logger = logi.NewDemonLogger("dual-protocol-server")
	}

	// Use default handler if none provided
	if handler == nil {
		handler = createDefaultHandler()
	}

	// Wrap handler to inject connection info into request context
	wrappedHandler := dualprotocol.WrapHandlerWithConnectionInfo(handler)

	// Create server configuration
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        wrappedHandler,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	return dualprotocol.NewServer(server, tlsConfig, logger), nil
}

// createDefaultHandler creates a simple default handler that shows protocol information
func createDefaultHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get connection info from context
		connInfo, hasConnInfo := dualprotocol.GetConnectionInfo(r)

		response := map[string]interface{}{
			"service":     "dual-protocol-server",
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		}

		if hasConnInfo {
			response["connection"] = map[string]interface{}{
				"protocol":     connInfo.Protocol,
				"is_tls":       connInfo.IsTLS,
				"tls_version":  connInfo.TLSVersion,
				"cipher_suite": connInfo.CipherSuite,
				"detected_at":  connInfo.DetectedAt.Format(time.RFC3339),
			}
		} else {
			// Fallback detection from request
			protocol := "HTTP"
			if r.TLS != nil {
				protocol = "HTTPS"
			}
			response["connection"] = map[string]interface{}{
				"protocol": protocol,
				"is_tls":   r.TLS != nil,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Protocol-Detected", response["connection"].(map[string]interface{})["protocol"].(string))

		if err := writeJSONResponse(w, response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}

// writeJSONResponse writes a JSON response to the ResponseWriter
func writeJSONResponse(w http.ResponseWriter, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
