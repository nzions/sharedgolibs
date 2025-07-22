package ca

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
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
					return
				}
				if httpsServer.TLSConfig == nil {
					t.Error("TLS config is nil")
					return
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

func TestUpdateTransportOnlyIf(t *testing.T) {
	// Save original transport to restore after tests
	originalTransport := http.DefaultClient.Transport

	tests := []struct {
		name             string
		setupEnv         func()
		wantError        bool
		wantLogMsg       bool
		expectTransport  bool // Whether transport should be modified
	}{
		{
			name: "No SGL_CA environment variable set",
			setupEnv: func() {
				os.Unsetenv("SGL_CA")
				os.Unsetenv("SGL_CA_API_KEY")
			},
			wantError:       false,
			wantLogMsg:      false,
			expectTransport: false, // No change to transport
		},
		{
			name: "Empty SGL_CA environment variable",
			setupEnv: func() {
				os.Setenv("SGL_CA", "")
				os.Unsetenv("SGL_CA_API_KEY")
			},
			wantError:       false,
			wantLogMsg:      false,
			expectTransport: false, // No change to transport
		},
		{
			name: "Invalid SGL_CA URL format",
			setupEnv: func() {
				os.Setenv("SGL_CA", "invalid-url")
				os.Unsetenv("SGL_CA_API_KEY")
			},
			wantError:       true,
			wantLogMsg:      false,
			expectTransport: false,
		},
		{
			name: "SGL_CA URL without scheme",
			setupEnv: func() {
				os.Setenv("SGL_CA", "localhost:8090")
				os.Unsetenv("SGL_CA_API_KEY")
			},
			wantError:       true,
			wantLogMsg:      false,
			expectTransport: false,
		},
		{
			name: "SGL_CA URL with unsupported scheme",
			setupEnv: func() {
				os.Setenv("SGL_CA", "ftp://localhost:8090")
				os.Unsetenv("SGL_CA_API_KEY")
			},
			wantError:       true,
			wantLogMsg:      false,
			expectTransport: false,
		},
		{
			name: "Valid SGL_CA but unreachable server",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:9999")
				os.Unsetenv("SGL_CA_API_KEY")
			},
			wantError:       true,
			wantLogMsg:      true,
			expectTransport: false,
		},
		{
			name: "Valid HTTPS SGL_CA but unreachable server",
			setupEnv: func() {
				os.Setenv("SGL_CA", "https://localhost:9999")
				os.Unsetenv("SGL_CA_API_KEY")
			},
			wantError:       true,
			wantLogMsg:      true,
			expectTransport: false,
		},
		{
			name: "Valid SGL_CA with API key but unreachable server",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:9998")
				os.Setenv("SGL_CA_API_KEY", "test-api-key")
			},
			wantError:       true,
			wantLogMsg:      true,
			expectTransport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset transport to original state before each test
			http.DefaultClient.Transport = originalTransport

			// Setup environment
			tt.setupEnv()
			defer func() {
				os.Unsetenv("SGL_CA")
				os.Unsetenv("SGL_CA_API_KEY")
				// Restore original transport after test
				http.DefaultClient.Transport = originalTransport
			}()

			// Store reference to transport before calling function
			transportBefore := http.DefaultClient.Transport

			// Call UpdateTransportOnlyIf
			err := UpdateTransportOnlyIf()

			// Check error expectation
			if (err != nil) != tt.wantError {
				t.Errorf("UpdateTransportOnlyIf() error = %v, wantError %v", err, tt.wantError)
			}

			// Check transport modification expectation
			transportAfter := http.DefaultClient.Transport
			transportChanged := transportBefore != transportAfter

			if tt.expectTransport && !transportChanged {
				t.Error("Expected transport to be modified, but it wasn't")
			} else if !tt.expectTransport && transportChanged {
				t.Error("Expected transport to remain unchanged, but it was modified")
			}

			// Verify transport configuration when it should be modified
			if tt.expectTransport && transportChanged {
				if transport, ok := transportAfter.(*http.Transport); ok {
					if transport.TLSClientConfig == nil {
						t.Error("Expected TLS config to be set when transport is modified")
					} else if transport.TLSClientConfig.RootCAs == nil {
						t.Error("Expected Root CAs to be set when transport is modified")
					}
				} else {
					t.Error("Expected modified transport to be *http.Transport type")
				}
			}
		})
	}
}

func TestUpdateTransportOnlyIfWithRealServer(t *testing.T) {
	// Save original transport to restore after test
	originalTransport := http.DefaultClient.Transport
	defer func() {
		http.DefaultClient.Transport = originalTransport
	}()

	// Start a test CA server
	server, err := NewServer(&ServerConfig{
		Port:     "18091",
		CAConfig: DefaultCAConfig(),
	})
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}

	// Start server in background
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name            string
		setupEnv        func()
		wantError       bool
		expectTransport bool
	}{
		{
			name: "Valid SGL_CA with running server",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:18091")
				os.Unsetenv("SGL_CA_API_KEY")
			},
			wantError:       false,
			expectTransport: true,
		},
		{
			name: "Valid SGL_CA with API key and running server",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:18091")
				os.Setenv("SGL_CA_API_KEY", "test-key")
			},
			wantError:       false,
			expectTransport: true,
		},
		{
			name: "Valid HTTPS SGL_CA with running server",
			setupEnv: func() {
				os.Setenv("SGL_CA", "https://localhost:18091")
				os.Unsetenv("SGL_CA_API_KEY")
			},
			wantError:       true, // HTTPS will fail because test server uses HTTP
			expectTransport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset transport before each test
			http.DefaultClient.Transport = originalTransport

			// Setup environment
			tt.setupEnv()
			defer func() {
				os.Unsetenv("SGL_CA")
				os.Unsetenv("SGL_CA_API_KEY")
			}()

			// Store reference to transport before calling function
			transportBefore := http.DefaultClient.Transport

			// Call UpdateTransportOnlyIf
			err := UpdateTransportOnlyIf()

			// Check error expectation
			if (err != nil) != tt.wantError {
				t.Errorf("UpdateTransportOnlyIf() error = %v, wantError %v", err, tt.wantError)
			}

			// Check transport modification expectation
			transportAfter := http.DefaultClient.Transport
			transportChanged := transportBefore != transportAfter

			if tt.expectTransport && !transportChanged {
				t.Error("Expected transport to be modified, but it wasn't")
			} else if !tt.expectTransport && transportChanged {
				t.Error("Expected transport to remain unchanged, but it was modified")
			}

			// For successful cases, verify the transport configuration
			if !tt.wantError && tt.expectTransport {
				if transport, ok := transportAfter.(*http.Transport); ok {
					if transport.TLSClientConfig == nil {
						t.Error("Expected TLS config to be set")
					} else if transport.TLSClientConfig.RootCAs == nil {
						t.Error("Expected Root CAs to be set")
					}
				} else {
					t.Error("Expected transport to be *http.Transport type")
				}
			}
		})
	}
}

func TestGetServiceCertificate(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		serviceIP   string
		domains     []string
		wantError   bool
	}{
		{
			name:        "Valid certificate request to unreachable server",
			serviceName: "test-client-service",
			serviceIP:   "192.168.1.100", 
			domains:     []string{"client.test.local"},
			wantError:   true, // Expected to fail because default server isn't running
		},
		{
			name:        "Empty service name",
			serviceName: "",
			serviceIP:   "192.168.1.100",
			domains:     []string{"test.local"},
			wantError:   true,
		},
		{
			name:        "Empty domains",
			serviceName: "test-service",
			serviceIP:   "192.168.1.100",
			domains:     []string{},
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GetServiceCertificate uses hardcoded localhost:8090
			resp, err := GetServiceCertificate(tt.serviceName, tt.serviceIP, tt.domains)

			if tt.wantError {
				if err == nil {
					t.Errorf("GetServiceCertificate() expected error but got none")
				}
				if resp != nil {
					t.Errorf("GetServiceCertificate() expected nil response on error")
				}
			} else {
				if err != nil {
					t.Errorf("GetServiceCertificate() unexpected error: %v", err)
				}
				if resp == nil {
					t.Errorf("GetServiceCertificate() returned nil response without error")
				}
			}
		})
	}
}

func TestUpdateTransportOnlyIfConditionalBehavior(t *testing.T) {
	// Save original transport to restore after test
	originalTransport := http.DefaultClient.Transport
	defer func() {
		http.DefaultClient.Transport = originalTransport
	}()

	t.Run("Function respects SGL_CA presence", func(t *testing.T) {
		// Test 1: No SGL_CA should not modify transport
		os.Unsetenv("SGL_CA")
		os.Unsetenv("SGL_CA_API_KEY")

		transportBefore := http.DefaultClient.Transport
		err := UpdateTransportOnlyIf()
		if err != nil {
			t.Errorf("UpdateTransportOnlyIf() should not error when SGL_CA unset: %v", err)
		}
		if http.DefaultClient.Transport != transportBefore {
			t.Error("Transport should not be modified when SGL_CA is unset")
		}

		// Test 2: Setting SGL_CA should attempt to modify transport (even if it fails due to bad URL)
		os.Setenv("SGL_CA", "http://invalid-test-url:99999")
		defer os.Unsetenv("SGL_CA")

		transportBefore = http.DefaultClient.Transport
		err = UpdateTransportOnlyIf()
		// We expect an error because the URL is unreachable, but the attempt should be made
		if err == nil {
			t.Error("UpdateTransportOnlyIf() should error with unreachable URL")
		}
		// Transport should remain unchanged on error
		if http.DefaultClient.Transport != transportBefore {
			t.Error("Transport should not be modified when update fails")
		}
	})

	t.Run("Function handles edge cases in environment variables", func(t *testing.T) {
		// Reset transport
		http.DefaultClient.Transport = originalTransport

		// Test with whitespace-only SGL_CA
		os.Setenv("SGL_CA", "   ")
		defer os.Unsetenv("SGL_CA")

		transportBefore := http.DefaultClient.Transport
		err := UpdateTransportOnlyIf()
		// Should error because whitespace-only URL fails validation
		if err == nil {
			t.Error("UpdateTransportOnlyIf() should error with whitespace-only SGL_CA because it fails URL validation")
		}
		if http.DefaultClient.Transport != transportBefore {
			t.Error("Transport should not be modified when URL validation fails")
		}
	})
}
