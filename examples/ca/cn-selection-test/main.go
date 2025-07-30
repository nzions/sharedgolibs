// Package main demonstrates CN selection logic with various SAN combinations
package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

func parseCertificate(certPEM string) *x509.Certificate {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		log.Fatal("Failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse certificate: %v", err)
	}

	return cert
}

func main() {
	fmt.Println("=== CN Selection Logic Test ===")

	// Create a new CA
	certAuthority, err := ca.NewCA(nil)
	if err != nil {
		log.Fatalf("Failed to create CA: %v", err)
	}

	testCases := []struct {
		name        string
		serviceName string
		sans        []string
		expectedCN  string
		shouldError bool
	}{
		{
			name:        "Mixed SANs (hostnames first)",
			serviceName: "web-service",
			sans:        []string{"api.example.com", "web.local", "192.168.1.100", "127.0.0.1"},
			expectedCN:  "api.example.com",
			shouldError: false,
		},
		{
			name:        "Mixed SANs (IP first)",
			serviceName: "api-service",
			sans:        []string{"10.0.0.1", "api.internal.com", "127.0.0.1"},
			expectedCN:  "api.internal.com", // Still first hostname
			shouldError: false,
		},
		{
			name:        "Only hostnames",
			serviceName: "auth-service",
			sans:        []string{"auth.example.com", "authentication.local"},
			expectedCN:  "auth.example.com",
			shouldError: false,
		},
		{
			name:        "Only IP addresses",
			serviceName: "internal-service",
			sans:        []string{"10.0.0.100", "172.16.1.50"},
			expectedCN:  "10.0.0.100", // First IP becomes CN
			shouldError: false,
		},
		{
			name:        "Single IP address",
			serviceName: "cache-service",
			sans:        []string{"192.168.1.200"},
			expectedCN:  "192.168.1.200",
			shouldError: false,
		},
		{
			name:        "Empty SANs",
			serviceName: "invalid-service",
			sans:        []string{},
			expectedCN:  "",
			shouldError: true,
		},
	}

	for i, tc := range testCases {
		fmt.Printf("\n%d. %s\n", i+1, tc.name)
		fmt.Printf("   Service: %s\n", tc.serviceName)
		fmt.Printf("   SANs: %v\n", tc.sans)

		req := ca.CertRequestV2{
			ServiceName: tc.serviceName,
			SANs:        tc.sans,
		}

		resp, err := certAuthority.IssueServiceCertificateV2(req)

		if tc.shouldError {
			if err != nil {
				fmt.Printf("   ✅ Expected error: %v\n", err)
			} else {
				fmt.Printf("   ❌ Expected error but got success\n")
			}
			continue
		}

		if err != nil {
			fmt.Printf("   ❌ Unexpected error: %v\n", err)
			continue
		}

		// Parse certificate and check CN
		cert := parseCertificate(resp.Certificate)
		actualCN := cert.Subject.CommonName

		if actualCN == tc.expectedCN {
			fmt.Printf("   ✅ CN: %s (correct)\n", actualCN)
		} else {
			fmt.Printf("   ❌ CN: %s (expected %s)\n", actualCN, tc.expectedCN)
		}

		// Show DNS names and IP addresses for context
		fmt.Printf("   DNS Names: %v\n", cert.DNSNames)
		ipStrs := make([]string, len(cert.IPAddresses))
		for j, ip := range cert.IPAddresses {
			ipStrs[j] = ip.String()
		}
		fmt.Printf("   IP Addresses: %v\n", ipStrs)
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("✨ CN Selection Rules:")
	fmt.Println("   1. First non-IP domain → CN")
	fmt.Println("   2. Only IPs → First IP as CN")
	fmt.Println("   3. Empty SANs → Error")
	fmt.Println("   4. No .local suffix added")
	fmt.Println("🎉 Test complete!")
}
