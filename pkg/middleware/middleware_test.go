// SPDX-License-Identifier: CC0-1.0

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nzions/sharedgolibs/pkg/logi"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Version should follow semantic versioning pattern (without 'v' prefix)
	if len(Version) < 5 {
		t.Errorf("Version %q should follow X.Y.Z format", Version)
	}
}

func TestWithCORS(t *testing.T) {
	handler := WithCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))

	// Test regular request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify all CORS headers are set correctly
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin to be '*', got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods to be 'GET, POST, PUT, DELETE, OPTIONS', got %s", w.Header().Get("Access-Control-Allow-Methods"))
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization, X-Requested-With, x-client-version, x-firebase-gmpid" {
		t.Errorf("Expected Access-Control-Allow-Headers to be 'Content-Type, Authorization, X-Requested-With, x-client-version, x-firebase-gmpid', got %s", w.Header().Get("Access-Control-Allow-Headers"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected Access-Control-Allow-Credentials to be 'true', got %s", w.Header().Get("Access-Control-Allow-Credentials"))
	}

	// Verify the request continues to the handler and response body is preserved
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got %s", w.Body.String())
	}

	// Test OPTIONS request (preflight)
	req = httptest.NewRequest("OPTIONS", "/test", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 for OPTIONS, got %d", w.Code)
	}

	// Verify CORS headers are also set for OPTIONS requests
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin to be '*' for OPTIONS, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods for OPTIONS, got %s", w.Header().Get("Access-Control-Allow-Methods"))
	}

	// Verify OPTIONS request doesn't call the next handler (body should be empty)
	if w.Body.String() != "" {
		t.Errorf("Expected empty body for OPTIONS request, got %s", w.Body.String())
	}
}

func TestWithCORSAllMethods(t *testing.T) {
	handler := WithCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("method: " + r.Method))
	}))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// All methods should get CORS headers
			if w.Header().Get("Access-Control-Allow-Origin") != "*" {
				t.Errorf("Expected Access-Control-Allow-Origin for %s, got %s", method, w.Header().Get("Access-Control-Allow-Origin"))
			}
			if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
				t.Errorf("Expected Access-Control-Allow-Methods for %s, got %s", method, w.Header().Get("Access-Control-Allow-Methods"))
			}

			// Non-OPTIONS methods should reach the handler
			if method != "OPTIONS" {
				if w.Code != http.StatusOK {
					t.Errorf("Expected status 200 for %s, got %d", method, w.Code)
				}
				expectedBody := "method: " + method
				if w.Body.String() != expectedBody {
					t.Errorf("Expected body '%s' for %s, got %s", expectedBody, method, w.Body.String())
				}
			}
		})
	}
}

func TestWithCORSPreflightOnly(t *testing.T) {
	handlerCalled := false
	handler := WithCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler called"))
	}))

	// Test that OPTIONS request doesn't call the next handler
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Expected OPTIONS request to not call the next handler")
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 for OPTIONS, got %d", w.Code)
	}
	if w.Body.String() != "" {
		t.Errorf("Expected empty body for OPTIONS, got %s", w.Body.String())
	}

	// Reset and test that non-OPTIONS request does call the handler
	handlerCalled = false
	req = httptest.NewRequest("POST", "/test", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected POST request to call the next handler")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for POST, got %d", w.Code)
	}
	if w.Body.String() != "handler called" {
		t.Errorf("Expected 'handler called' body for POST, got %s", w.Body.String())
	}
}

func TestWithLogging(t *testing.T) {
	// Test with BufferLogger for capturing output
	logger := logi.NewBufferLogger("test")

	handler := WithLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if len(logger.Messages) == 0 {
		t.Error("Expected at least one log message")
		return
	}

	logOutput := logger.Messages[0]
	if !strings.Contains(logOutput, "method=GET") || !strings.Contains(logOutput, "path=/test") {
		t.Errorf("Expected log entry with method and path, got %s", logOutput)
	}
}

func TestWithLoggingSlog(t *testing.T) {
	// Test with logi.DaemonLogger (which wraps slog)
	logger := logi.NewDemonLogger("test")

	handler := WithLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// The actual logging happens to stdout/discard in logi.DaemonLogger
	// This test mainly verifies that the interface works correctly
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestLogAndCORS(t *testing.T) {
	logger := logi.NewBufferLogger("test")

	handler := LogAndCORS(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should have all CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods, got %s", w.Header().Get("Access-Control-Allow-Methods"))
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization, X-Requested-With, x-client-version, x-firebase-gmpid" {
		t.Errorf("Expected Access-Control-Allow-Headers, got %s", w.Header().Get("Access-Control-Allow-Headers"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected Access-Control-Allow-Credentials 'true', got %s", w.Header().Get("Access-Control-Allow-Credentials"))
	}

	// Should have logged the request
	if len(logger.Messages) == 0 {
		t.Error("Expected at least one log message")
		return
	}

	logOutput := logger.Messages[0]
	if !strings.Contains(logOutput, "method=GET") || !strings.Contains(logOutput, "path=/test") {
		t.Errorf("Expected log entry with method and path, got %s", logOutput)
	}

	// Should have proper response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "test response" {
		t.Errorf("Expected body 'test response', got %s", w.Body.String())
	}

	// Test OPTIONS request with combined middleware
	logger = logi.NewBufferLogger("test") // Reset logger
	req = httptest.NewRequest("OPTIONS", "/test", nil)
	w = httptest.NewRecorder()
	handler = LogAndCORS(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 for OPTIONS in combined middleware, got %d", w.Code)
	}

	// OPTIONS should still be logged
	if len(logger.Messages) == 0 {
		t.Error("Expected OPTIONS request to be logged")
		return
	}

	logOutput = logger.Messages[0]
	if !strings.Contains(logOutput, "method=OPTIONS") {
		t.Errorf("Expected OPTIONS request to be logged, got %s", logOutput)
	}
}

func TestChain(t *testing.T) {
	var calls []string

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "middleware1")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "middleware2")
			next.ServeHTTP(w, r)
		})
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, "final")
	})

	handler := Chain(middleware1, middleware2)(finalHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	expected := []string{"middleware1", "middleware2", "final"}
	if len(calls) != len(expected) {
		t.Fatalf("Expected %d calls, got %d", len(expected), len(calls))
	}

	for i, call := range calls {
		if call != expected[i] {
			t.Errorf("Expected call %d to be %s, got %s", i, expected[i], call)
		}
	}
}

func TestWithGoogleMetadataFlavor(t *testing.T) {
	handler := WithGoogleMetadataFlavor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Metadata-Flavor") != "Google" {
		t.Errorf("Expected Metadata-Flavor header to be 'Google', got %s", w.Header().Get("Metadata-Flavor"))
	}

	if w.Header().Get("Server") != "Metadata Server for Google Compute Engine" {
		t.Errorf("Expected Server header to be 'Metadata Server for Google Compute Engine', got %s", w.Header().Get("Server"))
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got %s", w.Body.String())
	}
}

// SGLMux Tests

func TestNewSGLMux(t *testing.T) {
	// Test basic mux with no options
	mux := NewSGLMux()
	if mux == nil {
		t.Error("NewSGLMux should not return nil")
	}

	routes := mux.GetRoutes()
	if len(routes) != 0 {
		t.Errorf("New mux with no options should have 0 routes, got %d", len(routes))
	}

	// Test with health endpoint
	muxWithHealth := NewSGLMux(OptHealth())
	routesWithHealth := muxWithHealth.GetRoutes()
	if len(routesWithHealth) != 1 {
		t.Errorf("New mux with health should have 1 route (/health), got %d", len(routesWithHealth))
	}

	if routesWithHealth[0].Pattern != "GET /health" {
		t.Errorf("Expected health route pattern 'GET /health', got %s", routesWithHealth[0].Pattern)
	}
}

func TestSGLMux_HandleFunc(t *testing.T) {
	mux := NewSGLMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	routes := mux.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	if routes[0].Pattern != "/test" {
		t.Errorf("Expected pattern '/test', got %s", routes[0].Pattern)
	}

	// Test that it actually works
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got %s", w.Body.String())
	}
}

func TestSGLMux_Handle(t *testing.T) {
	mux := NewSGLMux()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	mux.Handle("/test", handler)

	routes := mux.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	if routes[0].Pattern != "/test" {
		t.Errorf("Expected pattern '/test', got %s", routes[0].Pattern)
	}
}

func TestSGLMux_Go122Patterns(t *testing.T) {
	mux := NewSGLMux()

	// Test Go 1.22+ patterns - let Go's ServeMux validate them
	validPatterns := []string{
		"/",
		"/test",
		"/api/v1/users",
		"GET /users",
		"POST /users",
		"PUT /users/{id}",
		"DELETE /users/{user_id}",
		"/files/{filename}",
		"/static/{filepath...}",
		"/users/{id}/posts/{post_id}",
	}

	for _, pattern := range validPatterns {
		t.Run(pattern, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Unexpected panic for valid pattern %q: %v", pattern, r)
				}
			}()
			mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {})
		})
	}

	routes := mux.GetRoutes()
	if len(routes) != len(validPatterns) {
		t.Errorf("Expected %d routes, got %d", len(validPatterns), len(routes))
	}
}

func TestSGLMux_InvalidPatterns(t *testing.T) {
	mux := NewSGLMux()

	// These should panic because Go's ServeMux will reject them
	invalidPatterns := []string{
		"",
		"no-slash",
		"/test/{}", // Go 1.22+ rejects empty wildcard names
	}

	for _, pattern := range invalidPatterns {
		t.Run(pattern, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected panic for invalid pattern %q", pattern)
				}
			}()
			mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {})
		})
	}
}

func TestSGLMux_GetRoutes(t *testing.T) {
	mux := NewSGLMux()

	mux.HandleFunc("/route1", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("GET /route2", func(w http.ResponseWriter, r *http.Request) {})

	routes := mux.GetRoutes()

	// Modify the returned slice - should not affect internal routes
	if len(routes) > 0 {
		routes[0].Pattern = "modified"
	}

	// Get routes again and verify original is unchanged
	routes2 := mux.GetRoutes()
	if len(routes2) > 0 && routes2[0].Pattern != "/route1" {
		t.Error("GetRoutes should return a copy, not the original slice")
	}
}

func TestSGLMux_HealthEndpoint(t *testing.T) {
	mux := NewSGLMux(OptHealth())

	// Test the /health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /health, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Parse the JSON response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	// Check that status is present
	status, ok := response["status"].(string)
	if !ok || status != "ok" {
		t.Error("Expected status 'ok' in health response")
	}

	// Check that routes array is present
	routes, ok := response["routes"].([]interface{})
	if !ok {
		t.Error("Expected routes array in health response")
	}

	// Should have 1 route: /health
	if len(routes) != 1 {
		t.Errorf("Expected 1 route in health response, got %d", len(routes))
	}
}

func TestSGLMux_HealthEndpointWithRoutes(t *testing.T) {
	mux := NewSGLMux(OptHealth())

	// Add some test routes
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {})

	// Test the /health endpoint now includes routes
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /health, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	// Check status
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}

	// Check routes are included
	routes, ok := response["routes"].([]any)
	if !ok {
		t.Errorf("Expected routes array in health response")
	}

	if len(routes) != 3 { // /health + 2 user routes
		t.Errorf("Expected 3 routes in health response, got %d", len(routes))
	}
}

func TestSGLMux_OptHealthWithCustomData(t *testing.T) {
	customData := map[string]any{
		"version":    "1.0.0",
		"build":      "abc123",
		"deployment": "production",
	}

	mux := NewSGLMux(OptHealth(customData))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /health, got %d", w.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	// Check status
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}

	// Check custom data is included
	data, ok := response["data"].(map[string]any)
	if !ok {
		t.Errorf("Expected data field in health response")
	}

	if data["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %v", data["version"])
	}

	if data["build"] != "abc123" {
		t.Errorf("Expected build 'abc123', got %v", data["build"])
	}

	if data["deployment"] != "production" {
		t.Errorf("Expected deployment 'production', got %v", data["deployment"])
	}
}

func TestSGLMux_WithMiddleware(t *testing.T) {
	// Test that users can wrap the mux with middleware in their main function
	logger := logi.NewBufferLogger("test")
	mux := NewSGLMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Wrap the entire mux with middleware
	wrappedMux := WithLogging(logger)(WithCORS(mux))

	// Test that it works with both CORS and logging
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedMux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS headers to be set when wrapped")
	}

	// Check that logging happened
	if len(logger.Messages) == 0 {
		t.Error("Expected logging middleware to generate logs when wrapped")
	}

	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got %s", w.Body.String())
	}
}

func TestSGLMux_OptHealth(t *testing.T) {
	// Test health endpoint only
	mux := NewSGLMux(OptHealth())
	routes := mux.GetRoutes()

	if len(routes) != 1 {
		t.Errorf("Expected 1 route with OptHealth, got %d", len(routes))
	}

	if routes[0].Pattern != "GET /health" {
		t.Errorf("Expected 'GET /health' route, got %s", routes[0].Pattern)
	}

	// Test that health endpoint works
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /health, got %d", w.Code)
	}
}
