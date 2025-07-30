// SPDX-License-Identifier: CC0-1.0

// Example demonstrating the difference between the standard CreateSecureHTTPSServer
// and the new CreateSecureDualProtocolServer functions.

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
	// Both functions follow the same pattern and automatically fetch certificates from CA

	// Example 1: Standard HTTPS-only server
	httpsServer, err := ca.CreateSecureHTTPSServer(
		"my-https-service",    // service name
		"127.0.0.1",           // service IP
		"8444",                // port
		[]string{"localhost"}, // domains
		createExampleHandler("HTTPS-only"),
	)
	if err != nil {
		log.Printf("Failed to create HTTPS server: %v", err)
	} else {
		fmt.Printf("Created HTTPS-only server: %+v\n", httpsServer)
	}

	// Example 2: New dual protocol server (handles both HTTP and HTTPS)
	dualServer, err := ca.CreateSecureDualProtocolServer(
		"my-dual-service",                     // service name
		"8443",                                // port
		[]string{"localhost", "127.0.0.1"},    // SANs (domains and IPs)
		createExampleHandler("Dual Protocol"), // handler
		nil,                                   // logger (nil = use default)
	)
	if err != nil {
		log.Printf("Failed to create dual protocol server: %v", err)
	} else {
		fmt.Printf("Created dual protocol server: %+v\n", dualServer)
	}

	fmt.Println("\nBoth servers:")
	fmt.Println("- Automatically fetch certificates from CA (SGL_CA environment variable)")
	fmt.Println("- Follow the same one-shot instantiation pattern")
	fmt.Println("- Are ready to call ListenAndServe() or ListenAndServeTLS()")

	fmt.Println("\nKey differences:")
	fmt.Println("- HTTPS-only: Requires ListenAndServeTLS(), only accepts HTTPS")
	fmt.Println("- Dual protocol: Uses ListenAndServe(), accepts both HTTP and HTTPS on same port")
}

func createExampleHandler(serverType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		protocol := "HTTP"
		if r.TLS != nil {
			protocol = "HTTPS"
		}

		fmt.Fprintf(w, "Hello from %s server via %s!\n", serverType, protocol)
	}
}
