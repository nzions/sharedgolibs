package middleware

import (
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	
	// Version should follow semantic versioning pattern
	if len(Version) < 5 || Version[0] != 'v' {
		t.Errorf("Version %q should follow vX.Y.Z format", Version)
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

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test OPTIONS request
	req = httptest.NewRequest("OPTIONS", "/test", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 for OPTIONS, got %d", w.Code)
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := LogAndCORS(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should have CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
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
