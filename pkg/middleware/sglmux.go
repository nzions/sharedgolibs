// SPDX-License-Identifier: CC0-1.0

package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
)

// SGLMux wraps http.ServeMux to record routes.
// It enforces Go 1.22+ mux patterns by letting ServeMux validate them.
type SGLMux struct {
	*http.ServeMux
	routes []Route
	mu     sync.RWMutex
}

// Route represents a registered route.
type Route struct {
	Pattern string `json:"pattern"`
}

// SGLMuxOption configures SGLMux.
type SGLMuxOption func(*SGLMux)

// OptHealth adds /health endpoint.
func OptHealth() SGLMuxOption {
	return func(m *SGLMux) {
		m.HandleFunc("GET /health", m.healthHandler)
	}
}

// OptServices adds /services endpoint.
func OptServices() SGLMuxOption {
	return func(m *SGLMux) {
		m.HandleFunc("GET /services", m.servicesHandler)
	}
}

// NewSGLMux creates a new SGLMux instance with the given options.
func NewSGLMux(opts ...SGLMuxOption) *SGLMux {
	mux := &SGLMux{
		ServeMux: http.NewServeMux(),
		routes:   make([]Route, 0),
	}

	// Apply options
	for _, opt := range opts {
		opt(mux)
	}

	return mux
}

// HandleFunc registers a handler function for the given pattern and records the route.
// Pattern validation is handled by Go's native ServeMux.
func (m *SGLMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	// Let Go's ServeMux validate the pattern - if it's invalid, it will panic
	m.ServeMux.HandleFunc(pattern, handler)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes = append(m.routes, Route{Pattern: pattern})
}

// Handle registers a handler for the given pattern and records the route.
func (m *SGLMux) Handle(pattern string, handler http.Handler) {
	m.ServeMux.Handle(pattern, handler)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes = append(m.routes, Route{Pattern: pattern})
}

// GetRoutes returns a copy of all registered routes.
func (m *SGLMux) GetRoutes() []Route {
	m.mu.RLock()
	defer m.mu.RUnlock()

	routes := make([]Route, len(m.routes))
	copy(routes, m.routes)
	return routes
}

// healthHandler handles the built-in /health endpoint.
func (m *SGLMux) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"routes": len(m.GetRoutes()),
	})
}

// servicesHandler handles the built-in /services endpoint.
func (m *SGLMux) servicesHandler(w http.ResponseWriter, r *http.Request) {
	routes := m.GetRoutes()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"routes": routes,
		"count":  len(routes),
	})
}
