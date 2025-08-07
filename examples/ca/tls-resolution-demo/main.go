// SPDX-License-Identifier: CC0-1.0

// Simple example demonstrating how to solve TLS certificate verification errors
// for public APIs like Stripe using the CA transport system.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
	fmt.Println("=== TLS Certificate Issue Resolution Example ===")
	fmt.Println()

	// Your original error:
	// tls: failed to verify certificate: x509: certificate signed by unknown authority

	// Solution 1: Use system certificate authorities (recommended for public APIs)
	fmt.Println("1. Creating HTTP client with system certificate authorities:")
	client, err := ca.CreateHTTPClientWithSystemAndCustomCAs(true, false)
	if err != nil {
		log.Fatalf("Failed to create HTTP client: %v", err)
	}

	// Test with a public API endpoint (using httpbin.org as example since we don't have Stripe API key)
	testURL := "https://httpbin.org/get"
	fmt.Printf("   Testing connection to: %s\n", testURL)

	resp, err := client.Get(testURL)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		resp.Body.Close()
		fmt.Printf("   ✅ Success: Status %s\n", resp.Status)
	}

	// Solution 2: Use system CAs + custom CA (for mixed environments)
	fmt.Println()
	fmt.Println("2. Creating HTTP client with system + custom CAs:")

	// This would use SGL_CA if set, otherwise just system CAs
	mixedClient, err := ca.CreateHTTPClientWithSystemAndCustomCAs(true, true)
	if err != nil {
		log.Fatalf("Failed to create mixed client: %v", err)
	}

	fmt.Printf("   Testing connection to: %s\n", testURL)
	resp, err = mixedClient.Get(testURL)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		resp.Body.Close()
		fmt.Printf("   ✅ Success: Status %s\n", resp.Status)
	}

	// Show environment variables status
	fmt.Println()
	fmt.Println("=== Environment Status ===")
	if sglCA := os.Getenv("SGL_CA"); sglCA != "" {
		fmt.Printf("SGL_CA: %s\n", sglCA)
	} else {
		fmt.Println("SGL_CA: Not set (using system CAs only)")
	}

	if apiKey := os.Getenv("SGL_CA_API_KEY"); apiKey != "" {
		fmt.Println("SGL_CA_API_KEY: Set")
	} else {
		fmt.Println("SGL_CA_API_KEY: Not set")
	}

	fmt.Println()
	fmt.Println("=== Recommendations ===")
	fmt.Println("For your Stripe API error:")
	fmt.Println("1. Use ca.CreateHTTPClientWithSystemAndCustomCAs(true, false)")
	fmt.Println("2. This includes system certificate authorities that trust public CAs")
	fmt.Println("3. For development with internal services, also set includeCustomCA=true")

	fmt.Println()
	fmt.Println("Example code:")
	fmt.Println(`
	// Create client that trusts system CAs (includes public CAs like Stripe uses)
	client, err := ca.CreateHTTPClientWithSystemAndCustomCAs(true, false)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	
	// Use this client for Stripe API calls
	req, _ := http.NewRequest("GET", "https://api.stripe.com/v1/account", nil)
	req.Header.Set("Authorization", "Bearer " + stripeAPIKey)
	resp, err := client.Do(req)
	`)
}
