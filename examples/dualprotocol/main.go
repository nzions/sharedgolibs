// SPDX-License-Identifier: CC0-1.0

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nzions/sharedgolibs/pkg/ca"
	"github.com/nzions/sharedgolibs/pkg/ca/dualprotocol"
	"github.com/nzions/sharedgolibs/pkg/logi"
)

func main() {
	// Create logger
	logger := logi.NewDemonLogger("dual-protocol-example")

	// Create custom handler that shows connection details
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get connection info from context
		connInfo, hasConnInfo := dualprotocol.GetConnectionInfo(r)

		fmt.Fprintf(w, "Dual Protocol Server Example\n")
		fmt.Fprintf(w, "===========================\n\n")
		fmt.Fprintf(w, "Request Details:\n")
		fmt.Fprintf(w, "  Method: %s\n", r.Method)
		fmt.Fprintf(w, "  Path: %s\n", r.URL.Path)
		fmt.Fprintf(w, "  Remote Addr: %s\n", r.RemoteAddr)
		fmt.Fprintf(w, "  User Agent: %s\n", r.UserAgent())
		fmt.Fprintf(w, "\n")

		if hasConnInfo {
			fmt.Fprintf(w, "Connection Information:\n")
			fmt.Fprintf(w, "  Protocol: %s\n", connInfo.Protocol)
			fmt.Fprintf(w, "  Is TLS: %t\n", connInfo.IsTLS)
			if connInfo.IsTLS {
				fmt.Fprintf(w, "  TLS Version: %s\n", connInfo.TLSVersion)
				fmt.Fprintf(w, "  Cipher Suite: %s\n", connInfo.CipherSuite)
			}
			fmt.Fprintf(w, "  Detected At: %s\n", connInfo.DetectedAt.Format(time.RFC3339))
		} else {
			// Fallback detection
			protocol := "HTTP"
			if r.TLS != nil {
				protocol = "HTTPS"
			}
			fmt.Fprintf(w, "Connection Information (fallback):\n")
			fmt.Fprintf(w, "  Protocol: %s\n", protocol)
			fmt.Fprintf(w, "  Is TLS: %t\n", r.TLS != nil)
		}

		fmt.Fprintf(w, "\nTry accessing this server with both HTTP and HTTPS!\n")
	})

	// Create dual protocol server with CA integration
	// Note: This requires SGL_CA environment variable to be set
	server, err := ca.CreateSecureDualProtocolServer(
		"dual-protocol-example", // service name
		"8443",                  // port
		[]string{ // SANs (Subject Alternative Names) - mix of hostnames and IPs
			"api.example.com", // First non-IP becomes the CommonName
			"localhost",       // DNS name
			"127.0.0.1",       // IP address (auto-detected)
			"192.168.1.100",   // IP address (auto-detected)
		},
		handler, // custom handler
		logger,  // logger
	)
	if err != nil {
		log.Fatalf("Failed to create dual protocol server: %v", err)
	}

	// Start server in background
	go func() {
		logger.Info("Starting dual protocol server", "addr", ":8443")
		logger.Info("Access via HTTP:  http://localhost:8443")
		logger.Info("Access via HTTPS: https://localhost:8443")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	} else {
		logger.Info("Server gracefully stopped")
	}
}
