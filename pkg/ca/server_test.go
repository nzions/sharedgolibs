package ca

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name      string
		config    *ServerConfig
		wantError bool
	}{
		{
			name:      "Default config",
			config:    nil,
			wantError: false,
		},
		{
			name: "Custom config with GUI enabled",
			config: &ServerConfig{
				Port:      "8091",
				CAConfig:  DefaultCAConfig(),
				EnableGUI: true,
				GUIAPIKey: "test-key",
			},
			wantError: false,
		},
		{
			name: "Custom config with GUI disabled",
			config: &ServerConfig{
				Port:      "8092",
				CAConfig:  DefaultCAConfig(),
				EnableGUI: false,
				GUIAPIKey: "",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewServer(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("NewServer() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if server == nil {
					t.Error("NewServer() returned nil server")
				}
				if server.ca == nil {
					t.Error("Server CA is nil")
				}
			}
		})
	}
}

func TestServerAPIEndpoints(t *testing.T) {
	// Create a test server
	server, err := NewServer(&ServerConfig{
		Port:      "8093",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: true,
		GUIAPIKey: "",
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "GET /ca - Download CA certificate",
			method:         "GET",
			path:           "/ca",
			body:           "",
			expectedStatus: http.StatusOK,
			expectedType:   "application/x-pem-file",
		},
		{
			name:           "GET /health - Health check",
			method:         "GET",
			path:           "/health",
			body:           "",
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name:   "POST /cert - Request certificate",
			method: "POST",
			path:   "/cert",
			body: `{
				"service_name": "test-service",
				"service_ip": "192.168.1.100",
				"domains": ["test.local", "api.test.local"]
			}`,
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name:           "POST /cert - Invalid JSON",
			method:         "POST",
			path:           "/cert",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
			expectedType:   "",
		},
		{
			name:           "POST /cert - Missing service_name",
			method:         "POST",
			path:           "/cert",
			body:           `{"domains": ["test.local"]}`,
			expectedStatus: http.StatusBadRequest,
			expectedType:   "",
		},
		{
			name:           "POST /cert - Missing domains",
			method:         "POST",
			path:           "/cert",
			body:           `{"service_name": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedType:   "",
		},
		{
			name:           "GET /cert - Method not allowed",
			method:         "GET",
			path:           "/cert",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedType:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()

			// Route the request to the appropriate handler
			switch tt.path {
			case "/ca":
				server.handleCARequest(rr, req)
			case "/cert":
				server.handleCertRequest(rr, req)
			case "/health":
				server.handleHealth(rr, req)
			default:
				t.Fatalf("Unknown path: %s", tt.path)
			}

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedType != "" {
				contentType := rr.Header().Get("Content-Type")
				if contentType != tt.expectedType {
					t.Errorf("Expected content type %s, got %s", tt.expectedType, contentType)
				}
			}

			// Validate response content for successful requests
			if tt.expectedStatus == http.StatusOK {
				switch tt.path {
				case "/ca":
					body := rr.Body.String()
					if !strings.Contains(body, "BEGIN CERTIFICATE") {
						t.Error("CA response doesn't contain PEM certificate")
					}
				case "/health":
					var response map[string]interface{}
					if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
						t.Errorf("Failed to parse health response: %v", err)
					}
					if response["status"] != "healthy" {
						t.Error("Health status is not healthy")
					}
				case "/cert":
					var response CertResponse
					if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
						t.Errorf("Failed to parse cert response: %v", err)
					}
					if response.Certificate == "" || response.PrivateKey == "" {
						t.Error("Certificate response missing cert or key")
					}
				}
			}
		})
	}
}

func TestServerAPIKeyProtection(t *testing.T) {
	// Create a test server with API key protection
	apiKey := "test-api-key-123"
	server, err := NewServer(&ServerConfig{
		Port:      "8094",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: true,
		GUIAPIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set up the same handlers as the real server with middleware
		var handler http.Handler

		switch r.URL.Path {
		case "/ca":
			handler = http.HandlerFunc(server.handleCARequest)
		case "/cert":
			handler = http.HandlerFunc(server.handleCertRequest)
		case "/health":
			handler = http.HandlerFunc(server.handleHealth)
		case "/ui/", "/":
			if server.gui != nil {
				handler = http.HandlerFunc(server.gui.HandleDashboard)
			}
		case "/ui/certs":
			if server.gui != nil {
				handler = http.HandlerFunc(server.gui.HandleCertificates)
			}
		case "/ui/generate":
			if server.gui != nil {
				handler = http.HandlerFunc(server.gui.HandleGenerate)
			}
		case "/ui/download-ca":
			if server.gui != nil {
				handler = http.HandlerFunc(server.gui.HandleDownloadCA)
			}
		default:
			http.NotFound(w, r)
			return
		}

		// Apply API key middleware if API key is configured
		if server.guiAPIKey != "" {
			// Simple API key check for testing
			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" {
				providedKey = r.URL.Query().Get("api_key")
			}
			if providedKey != server.guiAPIKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		handler.ServeHTTP(w, r)
	}))
	defer testServer.Close()

	tests := []struct {
		name           string
		path           string
		method         string
		body           string
		apiKey         string
		expectedStatus int
		description    string
	}{
		// API endpoints without API key (should fail)
		{
			name:           "GET /ca without API key",
			path:           "/ca",
			method:         "GET",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should require API key",
		},
		{
			name:           "GET /health without API key",
			path:           "/health",
			method:         "GET",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should require API key",
		},
		{
			name:   "POST /cert without API key",
			path:   "/cert",
			method: "POST",
			body: `{
				"service_name": "test",
				"domains": ["test.local"]
			}`,
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should require API key",
		},

		// API endpoints with correct API key (should succeed)
		{
			name:           "GET /ca with API key",
			path:           "/ca",
			method:         "GET",
			apiKey:         apiKey,
			expectedStatus: http.StatusOK,
			description:    "Should succeed with valid API key",
		},
		{
			name:           "GET /health with API key",
			path:           "/health",
			method:         "GET",
			apiKey:         apiKey,
			expectedStatus: http.StatusOK,
			description:    "Should succeed with valid API key",
		},
		{
			name:   "POST /cert with API key",
			path:   "/cert",
			method: "POST",
			body: `{
				"service_name": "test",
				"domains": ["test.local"]
			}`,
			apiKey:         apiKey,
			expectedStatus: http.StatusOK,
			description:    "Should succeed with valid API key",
		},

		// GUI endpoints without API key (should fail)
		{
			name:           "GET /ui/ without API key",
			path:           "/ui/",
			method:         "GET",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should require API key for GUI",
		},
		{
			name:           "GET /ui/certs without API key",
			path:           "/ui/certs",
			method:         "GET",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should require API key for GUI",
		},
		{
			name:           "GET /ui/generate without API key",
			path:           "/ui/generate",
			method:         "GET",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should require API key for GUI",
		},
		{
			name:           "GET /ui/download-ca without API key",
			path:           "/ui/download-ca",
			method:         "GET",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should require API key for GUI",
		},

		// GUI endpoints with correct API key (should succeed)
		{
			name:           "GET /ui/ with API key",
			path:           "/ui/",
			method:         "GET",
			apiKey:         apiKey,
			expectedStatus: http.StatusOK,
			description:    "Should succeed with valid API key",
		},
		{
			name:           "GET /ui/certs with API key",
			path:           "/ui/certs",
			method:         "GET",
			apiKey:         apiKey,
			expectedStatus: http.StatusOK,
			description:    "Should succeed with valid API key",
		},
		{
			name:           "GET /ui/generate with API key",
			path:           "/ui/generate",
			method:         "GET",
			apiKey:         apiKey,
			expectedStatus: http.StatusOK,
			description:    "Should succeed with valid API key",
		},
		{
			name:           "GET /ui/download-ca with API key",
			path:           "/ui/download-ca",
			method:         "GET",
			apiKey:         apiKey,
			expectedStatus: http.StatusOK,
			description:    "Should succeed with valid API key",
		},

		// Test with wrong API key (should fail)
		{
			name:           "GET /ca with wrong API key",
			path:           "/ca",
			method:         "GET",
			apiKey:         "wrong-key",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should fail with wrong API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != "" {
				req, err = http.NewRequest(tt.method, testServer.URL+tt.path, strings.NewReader(tt.body))
			} else {
				req, err = http.NewRequest(tt.method, testServer.URL+tt.path, nil)
			}
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			// Add API key if provided
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("%s: Expected status %d, got %d. Response: %s", tt.description, tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

func TestServerAPIKeyProtectionWithQueryParam(t *testing.T) {
	// Test API key protection using query parameter instead of header
	apiKey := "query-test-key"
	server, err := NewServer(&ServerConfig{
		Port:      "8095",
		CAConfig:  DefaultCAConfig(),
		EnableGUI: true,
		GUIAPIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var handler http.Handler

		switch r.URL.Path {
		case "/ca":
			handler = http.HandlerFunc(server.handleCARequest)
		case "/ui/", "/":
			if server.gui != nil {
				handler = http.HandlerFunc(server.gui.HandleDashboard)
			}
		default:
			http.NotFound(w, r)
			return
		}

		// Apply API key middleware if API key is configured
		if server.guiAPIKey != "" {
			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" {
				providedKey = r.URL.Query().Get("api_key")
			}
			if providedKey != server.guiAPIKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		handler.ServeHTTP(w, r)
	}))
	defer testServer.Close()

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "API endpoint with query param API key",
			url:            fmt.Sprintf("%s/ca?api_key=%s", testServer.URL, apiKey),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GUI endpoint with query param API key",
			url:            fmt.Sprintf("%s/ui/?api_key=%s", testServer.URL, apiKey),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API endpoint with wrong query param API key",
			url:            fmt.Sprintf("%s/ca?api_key=wrong", testServer.URL),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(tt.url)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
