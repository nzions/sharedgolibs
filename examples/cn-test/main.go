// SPDX-License-Identifier: CC0-1.0

package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
	// Create a new CA
	caInstance, err := ca.NewCA(nil)
	if err != nil {
		log.Fatalf("Failed to create CA: %v", err)
	}

	fmt.Println("=== Testing New CN Selection Logic ===")

	// Test case 1: First domain should become CN
	fmt.Println("\n1. First DNS name becomes CN:")
	testCN(caInstance, "test1", []string{"api.example.com", "127.0.0.1", "localhost"}, "api.example.com")

	// Test case 2: Skip IPs, use first DNS name
	fmt.Println("\n2. Skip IPs, use first DNS name:")
	testCN(caInstance, "test2", []string{"192.168.1.1", "10.0.0.1", "service.local", "api.local"}, "service.local")

	// Test case 3: Only IPs, fallback to serviceName.local
	fmt.Println("\n3. Only IPs, fallback to serviceName.local:")
	testCN(caInstance, "ip-only-service", []string{"192.168.1.1", "10.0.0.1", "127.0.0.1"}, "ip-only-service.local")

	// Test case 4: Empty domains, fallback to serviceName.local
	fmt.Println("\n4. Empty domains, fallback to serviceName.local:")
	testCN(caInstance, "empty-service", []string{}, "empty-service.local")

	// Test case 5: Mixed with localhost first
	fmt.Println("\n5. localhost first:")
	testCN(caInstance, "localhost-service", []string{"localhost", "api.example.com", "127.0.0.1"}, "localhost")
}

func testCN(caInstance *ca.CA, serviceName string, domains []string, expectedCN string) {
	certPEM, _, err := caInstance.GenerateCertificate(serviceName, "", domains)
	if err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}

	// Parse the certificate
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		log.Fatal("Failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse certificate: %v", err)
	}

	fmt.Printf("  Service: %s\n", serviceName)
	fmt.Printf("  Domains: %v\n", domains)
	fmt.Printf("  Expected CN: %s\n", expectedCN)
	fmt.Printf("  Actual CN: %s\n", cert.Subject.CommonName)
	fmt.Printf("  DNSNames: %v\n", cert.DNSNames)
	fmt.Printf("  IPAddresses: %v\n", cert.IPAddresses)

	if cert.Subject.CommonName == expectedCN {
		fmt.Printf("  ✅ PASS\n")
	} else {
		fmt.Printf("  ❌ FAIL - Expected %s, got %s\n", expectedCN, cert.Subject.CommonName)
	}
}
