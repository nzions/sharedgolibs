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
	routes     []Route
	mu         sync.RWMutex
	healthData any
}

// Route represents a registered route.
type Route struct {
	Pattern string `json:"pattern"`
}

// SGLMuxOption configures SGLMux.
type SGLMuxOption func(*SGLMux)

// OptHealth adds /health endpoint with optional additional data.
func OptHealth(data ...any) SGLMuxOption {
	return func(m *SGLMux) {
		if len(data) > 0 {
			m.healthData = data[0]
		}
		m.HandleFunc("GET /health", m.healthHandler)
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

// healthHandler handles GET /health requests with routes and optional custom data.
func (m *SGLMux) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Build response with routes and optional health data
	response := map[string]any{
		"status": "ok",
		"routes": m.GetRoutes(),
	}

	// Add custom health data if provided
	if m.healthData != nil {
		response["data"] = m.healthData
	}

	json.NewEncoder(w).Encode(response)
}
