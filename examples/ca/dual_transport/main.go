// SPDX-License-Identifier: CC0-1.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Check for required environment variables
	if os.Getenv("SGL_CA") == "" {
		slog.Error("SGL_CA environment variable is required")
		slog.Info("Example: export SGL_CA=http://localhost:8090")
		os.Exit(1)
	}

	// Create custom handler that shows which protocol was used
	handler := createServiceHandler()

	// Create secure dual protocol server using CA
	// This automatically fetches certificates from the CA server
	server, err := ca.CreateSecureDualProtocolServer(
		"dual-protocol-demo", // service name
		"8443",               // port
		[]string{"localhost", "dual-demo.local", "127.0.0.1"}, // SANs (domains and IPs)
		handler, // HTTP handler
		nil,     // logger (nil = use default)
	)
	if err != nil {
		slog.Error("Failed to create dual protocol server", "error", err)
		slog.Info("Make sure SGL_CA points to a running CA server")
		os.Exit(1)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		slog.Info("Starting CA-integrated dual protocol server",
			"addr", ":8443",
			"ca_url", os.Getenv("SGL_CA"))
		slog.Info("Server accepts both HTTP and HTTPS on the same port")
		slog.Info("Test with:")
		slog.Info("  HTTP:  curl http://localhost:8443/")
		slog.Info("  HTTPS: curl -k https://localhost:8443/")
		slog.Info("  Info:  curl http://localhost:8443/info")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal or context cancellation
	select {
	case sig := <-sigChan:
		slog.Info("Received shutdown signal", "signal", sig)
	case <-ctx.Done():
		slog.Info("Context cancelled")
	}

	// Graceful shutdown
	slog.Info("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", "error", err)
	} else {
		slog.Info("Server shutdown complete")
	}
}

// createServiceHandler creates a simple handler that demonstrates protocol detection
func createServiceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Detect protocol
		protocol := "HTTP"
		secure := false
		if r.TLS != nil {
			protocol = "HTTPS"
			secure = true
		}

		// Handle different endpoints
		switch r.URL.Path {
		case "/":
			handleRoot(w, r, protocol, secure)
		case "/health":
			handleHealth(w, r, protocol)
		case "/info":
			handleInfo(w, r, protocol, secure)
		default:
			handleNotFound(w, r, protocol)
		}

		// Log the request
		slog.Info("Request handled",
			"protocol", protocol,
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request, protocol string, secure bool) {
	statusIcon := "🔓"
	securityMsg := "insecure connection"
	if secure {
		statusIcon = "🔒"
		securityMsg = "secure connection"
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>CA-Integrated Dual Protocol Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .protocol { color: %s; font-weight: bold; font-size: 1.2em; }
        .secure { color: #28a745; }
        .insecure { color: #dc3545; }
        .endpoint { margin: 10px 0; padding: 10px; background: #f8f9fa; border-radius: 4px; }
        .code { background: #e9ecef; padding: 2px 6px; border-radius: 3px; font-family: monospace; }
        .icon { font-size: 1.5em; }
    </style>
</head>
<body>
    <div class="container">
        <h1><span class="icon">%s</span> CA-Integrated Dual Protocol Server</h1>
        <p>You are connecting via <span class="protocol %s">%s</span> (%s)</p>
        
        <h2>Available Endpoints:</h2>
        <div class="endpoint">
            <strong>/</strong> - This welcome page
        </div>
        <div class="endpoint">
            <strong>/health</strong> - Health check endpoint
        </div>
        <div class="endpoint">
            <strong>/info</strong> - Detailed connection information
        </div>
        
        <h2>Test Commands:</h2>
        <p>HTTP: <span class="code">curl http://localhost:8443/info</span></p>
        <p>HTTPS: <span class="code">curl -k https://localhost:8443/info</span></p>
        
        <h2>Features:</h2>
        <ul>
            <li>Single port accepts both HTTP and HTTPS</li>
            <li>Automatic protocol detection</li>
            <li>Certificates from CA server</li>
            <li>Same handlers for both protocols</li>
        </ul>
    </div>
</body>
</html>`,
		getProtocolColor(protocol),
		statusIcon,
		getSecurityClass(secure),
		protocol,
		securityMsg)
}

func handleHealth(w http.ResponseWriter, r *http.Request, protocol string) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
  "status": "healthy",
  "protocol": "%s",
  "service": "dual-protocol-demo",
  "timestamp": "%s",
  "ca_integrated": true
}`, protocol, time.Now().UTC().Format(time.RFC3339))
}

func handleInfo(w http.ResponseWriter, r *http.Request, protocol string, secure bool) {
	info := map[string]interface{}{
		"protocol":      protocol,
		"secure":        secure,
		"method":        r.Method,
		"path":          r.URL.Path,
		"remote_addr":   r.RemoteAddr,
		"user_agent":    r.UserAgent(),
		"headers":       r.Header,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"service":       "dual-protocol-demo",
		"ca_integrated": true,
	}

	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		cert := r.TLS.PeerCertificates[0]
		info["tls_info"] = map[string]interface{}{
			"version":      getTLSVersionName(r.TLS.Version),
			"cipher_suite": getTLSCipherSuiteName(r.TLS.CipherSuite),
			"server_name":  r.TLS.ServerName,
			"certificate": map[string]interface{}{
				"subject":    cert.Subject.String(),
				"issuer":     cert.Issuer.String(),
				"not_before": cert.NotBefore,
				"not_after":  cert.NotAfter,
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSONResponse(w, info); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request, protocol string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{
  "error": "Not Found",
  "path": "%s",
  "protocol": "%s",
  "available_endpoints": ["/", "/health", "/info"]
}`, r.URL.Path, protocol)
}

func getProtocolColor(protocol string) string {
	if protocol == "HTTPS" {
		return "#28a745"
	}
	return "#dc3545"
}

func getSecurityClass(secure bool) string {
	if secure {
		return "secure"
	}
	return "insecure"
}

func getTLSVersionName(version uint16) string {
	switch version {
	case 0x0301:
		return "TLS 1.0"
	case 0x0302:
		return "TLS 1.1"
	case 0x0303:
		return "TLS 1.2"
	case 0x0304:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (%x)", version)
	}
}

func getTLSCipherSuiteName(cipherSuite uint16) string {
	// Simplified cipher suite mapping
	switch cipherSuite {
	case 0x1301:
		return "TLS_AES_128_GCM_SHA256"
	case 0x1302:
		return "TLS_AES_256_GCM_SHA384"
	case 0x1303:
		return "TLS_CHACHA20_POLY1305_SHA256"
	default:
		return fmt.Sprintf("Unknown (%x)", cipherSuite)
	}
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
