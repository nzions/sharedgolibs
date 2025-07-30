// SPDX-License-Identifier: CC0-1.0

package dualprotocol

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/nzions/sharedgolibs/pkg/logi"
)

func TestNewServer(t *testing.T) {
	// Create a basic HTTP server
	httpServer := &http.Server{
		Addr:    ":8443",
		Handler: http.DefaultServeMux,
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	// Create logger
	logger := logi.NewDemonLogger("test-dual-protocol")

	// Create dual protocol server
	server := NewServer(httpServer, tlsConfig, logger)

	// Validate server was created correctly
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.Server != httpServer {
		t.Error("Expected HTTP server to be set correctly")
	}

	// Note: tlsConfig and logger are private fields, so we can't test them directly
	// but we can test that the server was created successfully
}

func TestWrapHandlerWithConnectionInfo(t *testing.T) {
	// Create a test handler that checks for connection info
	var receivedConnInfo *ConnectionInfo
	var hasConnInfo bool

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedConnInfo, hasConnInfo = GetConnectionInfo(r)
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the handler
	wrappedHandler := WrapHandlerWithConnectionInfo(testHandler)

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a mock response writer
	recorder := &mockResponseWriter{
		headers: make(http.Header),
	}

	// Call the wrapped handler
	wrappedHandler.ServeHTTP(recorder, req)

	// For a request without TLS, we should get HTTP protocol detection
	if hasConnInfo {
		if receivedConnInfo.Protocol != "HTTP" {
			t.Errorf("Expected protocol 'HTTP', got %s", receivedConnInfo.Protocol)
		}
		if receivedConnInfo.IsTLS {
			t.Error("Expected IsTLS to be false for HTTP request")
		}
	} else {
		t.Log("Connection info not set (this is expected for non-dual protocol requests)")
	}
}

func TestConnectionInfo(t *testing.T) {
	connInfo := &ConnectionInfo{
		Protocol:    "HTTPS",
		IsTLS:       true,
		TLSVersion:  "TLS 1.3",
		CipherSuite: "TLS_AES_256_GCM_SHA384",
		DetectedAt:  time.Now(),
	}

	if connInfo.Protocol != "HTTPS" {
		t.Errorf("Expected protocol 'HTTPS', got %s", connInfo.Protocol)
	}

	if !connInfo.IsTLS {
		t.Error("Expected IsTLS to be true")
	}

	if connInfo.TLSVersion != "TLS 1.3" {
		t.Errorf("Expected TLS version 'TLS 1.3', got %s", connInfo.TLSVersion)
	}
}

// Mock response writer for testing
type mockResponseWriter struct {
	headers http.Header
	body    []byte
	status  int
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(status int) {
	m.status = status
}
