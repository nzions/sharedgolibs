package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nzions/sharedgolibs/pkg/ca"
	"github.com/nzions/sharedgolibs/pkg/logi"
)

func main() {
	// Set up CA environment - use correct CA port
	os.Setenv("SGL_CA", "http://localhost:4200")

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test server with certificate details printing!"))
	})

	// Create logger
	logger := logi.NewDemonLogger("cert-test")

	// Test our enhanced CreateSecureDualProtocolServer with certificate details printing
	log.Println("Creating dual protocol server with certificate details printing...")
	server, err := ca.CreateSecureDualProtocolServer(
		"cert-test-service",
		"8445",
		[]string{"localhost", "127.0.0.1", "cert-test.local"},
		handler,
		logger,
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	go func() {
		log.Println("Starting server on :8445...")
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait a bit then test
	time.Sleep(2 * time.Second)

	// Test HTTP request
	resp, err := http.Get("http://localhost:8445/")
	if err != nil {
		log.Printf("HTTP test failed: %v", err)
	} else {
		resp.Body.Close()
		log.Printf("HTTP test successful: %d", resp.StatusCode)
	}

	// Test HTTPS request (skip verification for test)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err = client.Get("https://localhost:8445/")
	if err != nil {
		log.Printf("HTTPS test note: %v (expected for self-signed cert)", err)
	} else {
		resp.Body.Close()
		log.Printf("HTTPS test successful: %d", resp.StatusCode)
	}

	// Graceful shutdown
	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
