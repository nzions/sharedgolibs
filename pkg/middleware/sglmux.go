// SPDX-License-Identifier: CC0-1.0

package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
)

// HealthResponse represents the structure of health endpoint responses.
type HealthResponse struct {
	Status  string  `json:"status"`
	Routes  []Route `json:"routes,omitempty"`
	Version string  `json:"version,omitempty"`
	Data    any     `json:"data,omitempty"`
}

// SGLMux wraps http.ServeMux to record routes.
// It enforces Go 1.22+ mux patterns by letting ServeMux validate them.
type SGLMux struct {
	*http.ServeMux
	mu             sync.RWMutex
	healthResponse *HealthResponse
}

// Route represents a registered route.
type Route struct {
	Pattern string `json:"pattern"`
}

// SGLMuxOption configures SGLMux.
type SGLMuxOption func(*SGLMux)

// WithHealthData adds /health endpoint with optional additional data.
func WithHealthData(data any) SGLMuxOption {
	return func(m *SGLMux) {
		m.healthResponse.Data = data
		m.enableHealth()
	}
}

// WithVersion sets the version information to be included in health responses.
func WithVersion(version string) SGLMuxOption {
	return func(m *SGLMux) {
		m.healthResponse.Version = version
		m.enableHealth()
	}
}

// NewSGLMux creates a new SGLMux instance with the given options.
func NewSGLMux(opts ...SGLMuxOption) *SGLMux {
	mux := &SGLMux{
		ServeMux: http.NewServeMux(),
		healthResponse: &HealthResponse{
			Status: "ok",
			Routes: make([]Route, 0),
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(mux)
	}

	return mux
}

func (m *SGLMux) enableHealth() {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Check if health route already exists
	for _, route := range m.healthResponse.Routes {
		if route.Pattern == "GET /health" {
			return
		}
	}
	m.handleFunc("GET /health", m.healthHandler)
}

// HandleFunc registers a handler function for the given pattern and records the route.
// Pattern validation is handled by Go's native ServeMux.
func (m *SGLMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handleFunc(pattern, handler)
}

func (m *SGLMux) handleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.ServeMux.HandleFunc(pattern, handler)
	m.healthResponse.Routes = append(m.healthResponse.Routes, Route{Pattern: pattern})
}

// Handle registers a handler for the given pattern and records the route.
func (m *SGLMux) Handle(pattern string, handler http.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handle(pattern, handler)
}

func (m *SGLMux) handle(pattern string, handler http.Handler) {
	m.ServeMux.Handle(pattern, handler)
	m.healthResponse.Routes = append(m.healthResponse.Routes, Route{Pattern: pattern})
}

// GetRoutes returns a copy of all registered routes.
func (m *SGLMux) GetRoutes() []Route {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return a copy to prevent external modification of internal state
	routes := make([]Route, len(m.healthResponse.Routes))
	copy(routes, m.healthResponse.Routes)
	return routes
}

// healthHandler handles GET /health requests with routes and optional custom data.
func (m *SGLMux) healthHandler(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.healthResponse)
}
