// SPDX-License-Identifier: CC0-1.0

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

// StripeClient demonstrates how to create an HTTP client for Stripe API
// that handles TLS certificate verification issues
type StripeClient struct {
	client *http.Client
	apiKey string
}

// NewStripeClient creates a new Stripe client with proper TLS handling
func NewStripeClient(apiKey string, options ...ClientOption) *StripeClient {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// Start with secure defaults
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	sc := &StripeClient{
		client: client,
		apiKey: apiKey,
	}

	// Apply options
	for _, opt := range options {
		opt(sc)
	}

	return sc
}

// ClientOption defines configuration options for the StripeClient
type ClientOption func(*StripeClient)

// WithSystemCerts configures the client to use system certificate store
func WithSystemCerts() ClientOption {
	return func(sc *StripeClient) {
		transport := sc.client.Transport.(*http.Transport)

		// Get system certificate pool
		certPool, err := x509.SystemCertPool()
		if err != nil {
			log.Printf("Warning: Could not load system cert pool: %v", err)
			certPool = x509.NewCertPool()
		}

		transport.TLSClientConfig.RootCAs = certPool
	}
}

// WithCustomCA configures the client to trust additional CA certificates
func WithCustomCA(caCertPEM []byte) ClientOption {
	return func(sc *StripeClient) {
		transport := sc.client.Transport.(*http.Transport)

		// Start with existing pool or create new one
		certPool := transport.TLSClientConfig.RootCAs
		if certPool == nil {
			var err error
			certPool, err = x509.SystemCertPool()
			if err != nil {
				certPool = x509.NewCertPool()
			}
		}

		// Add the custom CA
		if !certPool.AppendCertsFromPEM(caCertPEM) {
			log.Printf("Warning: Failed to add custom CA certificate")
		}

		transport.TLSClientConfig.RootCAs = certPool
	}
}

// WithInsecureSkipVerify configures the client to skip TLS verification (INSECURE)
// Only use this for development/testing!
func WithInsecureSkipVerify() ClientOption {
	return func(sc *StripeClient) {
		transport := sc.client.Transport.(*http.Transport)
		transport.TLSClientConfig.InsecureSkipVerify = true
		log.Println("WARNING: TLS certificate verification disabled - INSECURE!")
	}
}

// WithCATransport configures the client to use the CA transport system
func WithCATransport() ClientOption {
	return func(sc *StripeClient) {
		// Try to update the global transport first
		if err := ca.UpdateTransportOnlyIf(); err != nil {
			log.Printf("Warning: Could not update CA transport: %v", err)
			return
		}

		// Use the updated default transport
		sc.client.Transport = http.DefaultTransport
	}
}

// GetAccount makes a request to Stripe's account endpoint
func (sc *StripeClient) GetAccount() error {
	req, err := http.NewRequest("GET", "https://api.stripe.com/v1/account", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+sc.apiKey)

	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Stripe API call successful: Status %s\n", resp.Status)
	return nil
}

func main() {
	// For demonstration - in real usage, get this from environment
	apiKey := os.Getenv("STRIPE_API_KEY")
	if apiKey == "" {
		apiKey = "sk_test_..." // placeholder
		fmt.Println("Note: Set STRIPE_API_KEY environment variable for real testing")
	}

	fmt.Println("=== Testing different TLS approaches for Stripe API ===\n")

	// Approach 1: Default client (may fail)
	fmt.Println("1. Testing with default settings:")
	client1 := NewStripeClient(apiKey)
	if err := client1.GetAccount(); err != nil {
		fmt.Printf("❌ Error: %v\n\n", err)
	}

	// Approach 2: With system certificates
	fmt.Println("2. Testing with system certificates:")
	client2 := NewStripeClient(apiKey, WithSystemCerts())
	if err := client2.GetAccount(); err != nil {
		fmt.Printf("❌ Error: %v\n\n", err)
	}

	// Approach 3: With CA transport (if configured)
	fmt.Println("3. Testing with CA transport system:")
	client3 := NewStripeClient(apiKey, WithCATransport())
	if err := client3.GetAccount(); err != nil {
		fmt.Printf("❌ Error: %v\n\n", err)
	}

	// Approach 4: Insecure (for development only)
	fmt.Println("4. Testing with insecure skip verify (DEVELOPMENT ONLY):")
	client4 := NewStripeClient(apiKey, WithInsecureSkipVerify())
	if err := client4.GetAccount(); err != nil {
		fmt.Printf("❌ Error: %v\n\n", err)
	}

	fmt.Println("=== Recommendations ===")
	fmt.Println("1. For production: Use WithSystemCerts() - this should work for public APIs like Stripe")
	fmt.Println("2. For development with custom CAs: Use WithCATransport() with SGL_CA configured")
	fmt.Println("3. For debugging only: Use WithInsecureSkipVerify() (NEVER in production)")
	fmt.Println("4. For custom corporate environments: Use WithCustomCA() with your company's CA")
}
