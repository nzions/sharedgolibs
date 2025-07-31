// SPDX-License-Identifier: CC0-1.0

package middleware

import (
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	// Test with log.Logger
	var buf strings.Builder
	logger := log.New(&buf, "", 0)

	handler := WithLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !strings.Contains(buf.String(), "GET /test") {
		t.Errorf("Expected log entry, got %s", buf.String())
	}
}

func TestWithLoggingSlog(t *testing.T) {
	// Test with slog.Logger
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := WithLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "method=GET") || !strings.Contains(logOutput, "path=/test") {
		t.Errorf("Expected slog entry with method and path, got %s", logOutput)
	}
}

func TestLogAndCORS(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, nil))

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
	logOutput := buf.String()
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
	buf.Reset()
	req = httptest.NewRequest("OPTIONS", "/test", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 for OPTIONS in combined middleware, got %d", w.Code)
	}

	// OPTIONS should still be logged
	logOutput = buf.String()
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
