// SPDX-License-Identifier: CC0-1.0

package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestNewSGLMux(t *testing.T) {
	// Test basic mux with no options
	mux := NewSGLMux()
	if mux == nil {
		t.Fatal("NewSGLMux should not return nil")
	}

	if mux.ServeMux == nil {
		t.Error("NewSGLMux should initialize ServeMux")
	}

	if mux.healthResponse == nil {
		t.Error("NewSGLMux should initialize healthResponse")
	}

	if mux.healthResponse.Status != "ok" {
		t.Errorf("Expected status 'ok', got %s", mux.healthResponse.Status)
	}

	routes := mux.GetRoutes()
	if len(routes) != 0 {
		t.Errorf("New mux with no options should have 0 routes, got %d", len(routes))
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

func TestSGLMux_WithHealthData(t *testing.T) {
	customData := map[string]any{
		"version":    "1.0.0",
		"build":      "abc123",
		"deployment": "production",
	}

	mux := NewSGLMux(WithHealthData(customData))

	// Check that health endpoint was automatically added
	routes := mux.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route with WithHealthData, got %d", len(routes))
	}

	if routes[0].Pattern != "GET /health" {
		t.Errorf("Expected 'GET /health' route, got %s", routes[0].Pattern)
	}

	// Test the health endpoint
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

func TestSGLMux_WithVersion(t *testing.T) {
	version := "v1.2.3"
	mux := NewSGLMux(WithVersion(version))

	// Check that health endpoint was automatically added
	routes := mux.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route with WithVersion, got %d", len(routes))
	}

	// Test the health endpoint includes version
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

	// Check version is included
	if response["version"] != version {
		t.Errorf("Expected version %s, got %v", version, response["version"])
	}
}

func TestSGLMux_MultipleOptions(t *testing.T) {
	customData := map[string]string{"service": "test"}
	version := "v1.0.0"

	mux := NewSGLMux(
		WithHealthData(customData),
		WithVersion(version),
	)

	// Should only have one health route, not duplicate
	routes := mux.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route with multiple options, got %d", len(routes))
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	// Both version and data should be present
	if response["version"] != version {
		t.Errorf("Expected version %s, got %v", version, response["version"])
	}

	data, ok := response["data"].(map[string]any)
	if !ok {
		t.Errorf("Expected data field in health response")
	}

	if data["service"] != "test" {
		t.Errorf("Expected service 'test', got %v", data["service"])
	}
}

func TestSGLMux_HealthWithRoutes(t *testing.T) {
	mux := NewSGLMux(WithHealthData(nil))

	// Add some test routes
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle("/api/static", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Test the /health endpoint includes all routes
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

	// Check routes are included
	routes, ok := response["routes"].([]any)
	if !ok {
		t.Errorf("Expected routes array in health response")
	}

	if len(routes) != 4 { // /health + 3 user routes
		t.Errorf("Expected 4 routes in health response, got %d", len(routes))
	}

	// Verify specific routes exist
	routePatterns := make([]string, len(routes))
	for i, route := range routes {
		routeMap, ok := route.(map[string]any)
		if !ok {
			t.Errorf("Expected route to be a map")
			continue
		}
		routePatterns[i] = routeMap["pattern"].(string)
	}

	expectedPatterns := []string{"GET /health", "/api/users", "POST /api/users", "/api/static"}
	for _, expected := range expectedPatterns {
		found := false
		for _, actual := range routePatterns {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find route pattern %s", expected)
		}
	}
}

// Test concurrent access to ensure thread safety
func TestSGLMux_ConcurrentAccess(t *testing.T) {
	mux := NewSGLMux()

	var wg sync.WaitGroup
	numGoroutines := 10
	routesPerGoroutine := 5

	// Concurrently add routes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < routesPerGoroutine; j++ {
				pattern := fmt.Sprintf("/route%d_%d", base, j)
				mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {})
			}
		}(i)
	}

	// Concurrently read routes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < routesPerGoroutine; j++ {
				routes := mux.GetRoutes()
				_ = routes // Just access, don't validate count since it's changing
			}
		}()
	}

	wg.Wait()

	// Final verification
	routes := mux.GetRoutes()
	expectedCount := numGoroutines * routesPerGoroutine
	if len(routes) != expectedCount {
		t.Errorf("Expected %d routes after concurrent access, got %d", expectedCount, len(routes))
	}
}

func TestSGLMux_EnableHealthDuplicatePrevention(t *testing.T) {
	mux := NewSGLMux()

	// Call enableHealth multiple times
	mux.enableHealth()
	mux.enableHealth()
	mux.enableHealth()

	routes := mux.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 health route despite multiple enableHealth calls, got %d", len(routes))
	}

	if routes[0].Pattern != "GET /health" {
		t.Errorf("Expected 'GET /health' route, got %s", routes[0].Pattern)
	}
}
