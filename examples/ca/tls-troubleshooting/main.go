// SPDX-License-Identifier: CC0-1.0

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

// Example showing different ways to handle TLS certificate issues
func main() {
	// Option 1: Use CA transport to add custom CA (if you have one)
	// This requires SGL_CA environment variable pointing to your CA server
	if err := ca.UpdateTransportOnlyIf(); err != nil {
		log.Printf("Failed to update transport with custom CA: %v", err)
	}

	// Option 2: Create a custom HTTP client that skips verification (INSECURE - only for development)
	insecureClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Option 3: Create a custom HTTP client with specific root CAs
	// This is the most secure approach if you know the specific CA you want to trust
	customClient, err := createCustomClientWithCA()
	if err != nil {
		log.Printf("Failed to create custom client: %v", err)
	}

	// Option 4: Add system certificate store + custom CAs
	systemPlusCustomClient, err := createSystemPlusCustomClient()
	if err != nil {
		log.Printf("Failed to create system+custom client: %v", err)
	}

	// Test the different approaches
	testURL := "https://api.stripe.com/v1/account"

	fmt.Println("Testing with default client (may fail):")
	testClient(http.DefaultClient, testURL)

	fmt.Println("\nTesting with insecure client (will work but insecure):")
	testClient(insecureClient, testURL)

	if customClient != nil {
		fmt.Println("\nTesting with custom CA client:")
		testClient(customClient, testURL)
	}

	if systemPlusCustomClient != nil {
		fmt.Println("\nTesting with system+custom CA client:")
		testClient(systemPlusCustomClient, testURL)
	}
}

// createCustomClientWithCA creates an HTTP client that trusts only specific CAs
func createCustomClientWithCA() (*http.Client, error) {
	// This would typically load your custom CA certificate
	// For demonstration, we'll create an empty pool (will fail for public sites)
	certPool := x509.NewCertPool()

	// You would add your custom CA like this:
	// if !certPool.AppendCertsFromPEM(yourCACertPEM) {
	//     return nil, fmt.Errorf("failed to add CA certificate")
	// }

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
	}, nil
}

// createSystemPlusCustomClient creates an HTTP client that trusts system CAs + custom CAs
func createSystemPlusCustomClient() (*http.Client, error) {
	// Start with system certificate pool
	certPool, err := x509.SystemCertPool()
	if err != nil {
		// Fallback to empty pool if system pool unavailable
		certPool = x509.NewCertPool()
	}

	// Add your custom CA certificates here
	// if !certPool.AppendCertsFromPEM(yourCACertPEM) {
	//     return nil, fmt.Errorf("failed to add custom CA certificate")
	// }

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
	}, nil
}

// testClient tests an HTTP client against a URL
func testClient(client *http.Client, url string) {
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Success: Status %s\n", resp.Status)
}
