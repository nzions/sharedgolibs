package ca

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUpdateTransport(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalAPIKey := os.Getenv("SGL_CA_API_KEY")
	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("SGL_CA_API_KEY", originalAPIKey)
	}()

	// Create a test CA server
	server, err := NewServer(&ServerConfig{
		Port:      "8096",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: false,
		GUIAPIKey: "",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ca" {
			server.handleCARequest(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	tests := []struct {
		name      string
		caURL     string
		apiKey    string
		wantError bool
	}{
		{
			name:      "Valid CA URL without API key",
			caURL:     testServer.URL,
			apiKey:    "",
			wantError: false,
		},
		{
			name:      "Invalid CA URL",
			caURL:     "http://nonexistent.local",
			apiKey:    "",
			wantError: true,
		},
		{
			name:      "Empty CA URL",
			caURL:     "",
			apiKey:    "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for test
			os.Setenv("SGL_CA", tt.caURL)
			os.Setenv("SGL_CA_API_KEY", tt.apiKey)

			err := UpdateTransport()

			if (err != nil) != tt.wantError {
				t.Errorf("UpdateTransport() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				// Verify that the default HTTP client was updated
				transport, ok := http.DefaultClient.Transport.(*http.Transport)
				if !ok {
					t.Error("Default client transport was not updated")
				} else if transport.TLSClientConfig == nil {
					t.Error("TLS config was not set")
				} else if transport.TLSClientConfig.RootCAs == nil {
					t.Error("Root CAs were not set")
				}
			}
		})
	}
}

func TestUpdateTransportWithAPIKey(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalAPIKey := os.Getenv("SGL_CA_API_KEY")
	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("SGL_CA_API_KEY", originalAPIKey)
	}()

	// Create a test CA server with API key protection
	apiKey := "transport-test-key"
	server, err := NewServer(&ServerConfig{
		Port:      "8097",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: false,
		GUIAPIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start test server with API key middleware
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ca" {
			// Check API key
			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" {
				providedKey = r.URL.Query().Get("api_key")
			}
			if providedKey != apiKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			server.handleCARequest(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	tests := []struct {
		name      string
		caURL     string
		apiKey    string
		wantError bool
	}{
		{
			name:      "Valid CA URL with correct API key",
			caURL:     testServer.URL,
			apiKey:    apiKey,
			wantError: false,
		},
		{
			name:      "Valid CA URL with wrong API key",
			caURL:     testServer.URL,
			apiKey:    "wrong-key",
			wantError: true,
		},
		{
			name:      "Valid CA URL without API key (should fail)",
			caURL:     testServer.URL,
			apiKey:    "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for test
			os.Setenv("SGL_CA", tt.caURL)
			os.Setenv("SGL_CA_API_KEY", tt.apiKey)

			err := UpdateTransport()

			if (err != nil) != tt.wantError {
				t.Errorf("UpdateTransport() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRequestCertificate(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalAPIKey := os.Getenv("SGL_CA_API_KEY")
	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("SGL_CA_API_KEY", originalAPIKey)
	}()

	// Create a test CA server
	server, err := NewServer(&ServerConfig{
		Port:      "8098",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: false,
		GUIAPIKey: "",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ca":
			server.handleCARequest(w, r)
		case "/cert":
			server.handleCertRequest(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	tests := []struct {
		name        string
		caURL       string
		serviceName string
		serviceIP   string
		domains     []string
		apiKey      string
		wantError   bool
	}{
		{
			name:        "Valid certificate request",
			caURL:       testServer.URL,
			serviceName: "test-service",
			serviceIP:   "192.168.1.100",
			domains:     []string{"test.local", "api.test.local"},
			apiKey:      "",
			wantError:   false,
		},
		{
			name:        "Invalid CA URL",
			caURL:       "http://nonexistent.local",
			serviceName: "test-service",
			serviceIP:   "192.168.1.100",
			domains:     []string{"test.local"},
			apiKey:      "",
			wantError:   true,
		},
		{
			name:        "Empty service name",
			caURL:       testServer.URL,
			serviceName: "",
			serviceIP:   "192.168.1.100",
			domains:     []string{"test.local"},
			apiKey:      "",
			wantError:   true,
		},
		{
			name:        "Empty domains",
			caURL:       testServer.URL,
			serviceName: "test-service",
			serviceIP:   "192.168.1.100",
			domains:     []string{},
			apiKey:      "",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for test
			os.Setenv("SGL_CA", tt.caURL)
			os.Setenv("SGL_CA_API_KEY", tt.apiKey)

			certResp, err := RequestCertificate(tt.serviceName, tt.serviceIP, tt.domains)

			if (err != nil) != tt.wantError {
				t.Errorf("RequestCertificate() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if certResp == nil {
					t.Error("Certificate response is nil")
					return
				}
				if certResp.Certificate == "" {
					t.Error("Certificate PEM is empty")
				}
				if certResp.PrivateKey == "" {
					t.Error("Private key PEM is empty")
				}
				if !strings.Contains(certResp.Certificate, "BEGIN CERTIFICATE") {
					t.Error("Certificate PEM doesn't contain valid certificate")
				}
				if !strings.Contains(certResp.PrivateKey, "BEGIN") || (!strings.Contains(certResp.PrivateKey, "PRIVATE KEY") && !strings.Contains(certResp.PrivateKey, "RSA PRIVATE KEY")) {
					t.Error("Private key PEM doesn't contain valid private key")
				}
			}
		})
	}
}

func TestRequestCertificateWithAPIKey(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalAPIKey := os.Getenv("SGL_CA_API_KEY")
	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("SGL_CA_API_KEY", originalAPIKey)
	}()

	// Create a test CA server with API key protection
	apiKey := "cert-request-test-key"
	server, err := NewServer(&ServerConfig{
		Port:      "8099",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: false,
		GUIAPIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start test server with API key middleware
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check API key for all endpoints
		providedKey := r.Header.Get("X-API-Key")
		if providedKey == "" {
			providedKey = r.URL.Query().Get("api_key")
		}
		if providedKey != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/ca":
			server.handleCARequest(w, r)
		case "/cert":
			server.handleCertRequest(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	tests := []struct {
		name        string
		caURL       string
		serviceName string
		serviceIP   string
		domains     []string
		apiKey      string
		wantError   bool
	}{
		{
			name:        "Valid request with correct API key",
			caURL:       testServer.URL,
			serviceName: "test-service",
			serviceIP:   "192.168.1.100",
			domains:     []string{"test.local"},
			apiKey:      apiKey,
			wantError:   false,
		},
		{
			name:        "Valid request with wrong API key",
			caURL:       testServer.URL,
			serviceName: "test-service",
			serviceIP:   "192.168.1.100",
			domains:     []string{"test.local"},
			apiKey:      "wrong-key",
			wantError:   true,
		},
		{
			name:        "Valid request without API key",
			caURL:       testServer.URL,
			serviceName: "test-service",
			serviceIP:   "192.168.1.100",
			domains:     []string{"test.local"},
			apiKey:      "",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for test
			os.Setenv("SGL_CA", tt.caURL)
			os.Setenv("SGL_CA_API_KEY", tt.apiKey)

			certResp, err := RequestCertificate(tt.serviceName, tt.serviceIP, tt.domains)

			if (err != nil) != tt.wantError {
				t.Errorf("RequestCertificate() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if certResp == nil || certResp.Certificate == "" || certResp.PrivateKey == "" {
					t.Error("Certificate response is nil or missing cert/key")
				}
			}
		})
	}
}

func TestCreateSecureHTTPSServer(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalAPIKey := os.Getenv("SGL_CA_API_KEY")
	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("SGL_CA_API_KEY", originalAPIKey)
	}()

	// Create a test CA server
	server, err := NewServer(&ServerConfig{
		Port:      "8100",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: false,
		GUIAPIKey: "",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ca":
			server.handleCARequest(w, r)
		case "/cert":
			server.handleCertRequest(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	tests := []struct {
		name        string
		caURL       string
		serviceName string
		serviceIP   string
		domains     []string
		port        string
		handler     http.Handler
		apiKey      string
		wantError   bool
	}{
		{
			name:        "Valid HTTPS server creation",
			caURL:       testServer.URL,
			serviceName: "test-https-service",
			serviceIP:   "127.0.0.1",
			domains:     []string{"localhost", "127.0.0.1"},
			port:        "0", // Use random port for testing
			handler:     http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) }),
			apiKey:      "",
			wantError:   false,
		},
		{
			name:        "Invalid CA URL",
			caURL:       "http://nonexistent.local",
			serviceName: "test-service",
			serviceIP:   "127.0.0.1",
			domains:     []string{"localhost"},
			port:        "0",
			handler:     http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) }),
			apiKey:      "",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for test
			os.Setenv("SGL_CA", tt.caURL)
			os.Setenv("SGL_CA_API_KEY", tt.apiKey)

			httpsServer, err := CreateSecureHTTPSServer(tt.serviceName, tt.serviceIP, tt.port, tt.domains, tt.handler)

			if (err != nil) != tt.wantError {
				t.Errorf("CreateSecureHTTPSServer() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if httpsServer == nil {
					t.Error("HTTPS server is nil")
				}
				if httpsServer.TLSConfig == nil {
					t.Error("TLS config is nil")
				}
				if len(httpsServer.TLSConfig.Certificates) == 0 {
					t.Error("No TLS certificates configured")
				}
			}
		})
	}
}

func TestEnvironmentVariableDetection(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalAPIKey := os.Getenv("SGL_CA_API_KEY")
	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("SGL_CA_API_KEY", originalAPIKey)
	}()

	// Create a test CA server
	server, err := NewServer(&ServerConfig{
		Port:      "8101",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: false,
		GUIAPIKey: "",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ca" {
			server.handleCARequest(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	// Set environment variables
	os.Setenv("SGL_CA", testServer.URL)
	os.Setenv("SGL_CA_API_KEY", "env-test-key")

	err = UpdateTransport()
	if err != nil {
		t.Fatalf("UpdateTransport() with env vars failed: %v", err)
	}

	// Verify that default transport was updated
	transport, ok := http.DefaultClient.Transport.(*http.Transport)
	if !ok {
		t.Error("Default client transport was not updated")
	} else if transport.TLSClientConfig == nil {
		t.Error("TLS config was not set")
	} else if transport.TLSClientConfig.RootCAs == nil {
		t.Error("Root CAs were not set")
	}
}

func TestTransportConvenienceMethodsIntegration(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalAPIKey := os.Getenv("SGL_CA_API_KEY")
	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("SGL_CA_API_KEY", originalAPIKey)
	}()

	// Create a test CA server with API key
	apiKey := "integration-test-key"
	server, err := NewServer(&ServerConfig{
		Port:      "8102",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: false,
		GUIAPIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start test server with API key middleware
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check API key
		providedKey := r.Header.Get("X-API-Key")
		if providedKey == "" {
			providedKey = r.URL.Query().Get("api_key")
		}
		if providedKey != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/ca":
			server.handleCARequest(w, r)
		case "/cert":
			server.handleCertRequest(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	t.Run("Full integration test", func(t *testing.T) {
		// Set environment variables
		os.Setenv("SGL_CA", testServer.URL)
		os.Setenv("SGL_CA_API_KEY", apiKey)

		// Test UpdateTransport
		err := UpdateTransport()
		if err != nil {
			t.Fatalf("UpdateTransport failed: %v", err)
		}

		// Test RequestCertificate
		certResp, err := RequestCertificate("integration-test-service", "127.0.0.1", []string{"localhost", "integration.test"})
		if err != nil {
			t.Fatalf("RequestCertificate failed: %v", err)
		}

		if certResp == nil || certResp.Certificate == "" || certResp.PrivateKey == "" {
			t.Fatal("Certificate response is nil or missing cert/key")
		}

		// Test CreateSecureHTTPSServer
		httpsServer, err := CreateSecureHTTPSServer("test-https", "127.0.0.1", "0", []string{"localhost"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Integration test OK"))
		}))
		if err != nil {
			t.Fatalf("CreateSecureHTTPSServer failed: %v", err)
		}

		if httpsServer == nil {
			t.Fatal("HTTPS server is nil")
		}

		// Verify TLS configuration
		if httpsServer.TLSConfig == nil {
			t.Fatal("TLS config is nil")
		}
		if len(httpsServer.TLSConfig.Certificates) == 0 {
			t.Fatal("No TLS certificates configured")
		}
	})
}
