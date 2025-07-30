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

	// Test case 1: Mixed IP addresses and hostnames
	fmt.Println("=== Test Case 1: Mixed IPs and Hostnames ===")
	serviceName := "mixed-service"
	serviceIP := "192.168.1.100"
	domains := []string{
		"api.example.com", // DNS name
		"service.local",   // DNS name
		"127.0.0.1",       // IPv4 address
		"::1",             // IPv6 address
		"10.0.0.1",        // Another IPv4
		"*.example.com",   // Wildcard DNS name
	}

	certPEM, _, err := caInstance.GenerateCertificate(serviceName, serviceIP, domains)
	if err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}

	printCertificateDetails("Mixed Test", certPEM)

	// Test case 2: Only DNS names
	fmt.Println("\n=== Test Case 2: Only DNS Names ===")
	serviceName2 := "dns-only-service"
	serviceIP2 := "" // No IP
	domains2 := []string{
		"app.company.com",
		"api.company.com",
		"*.company.com",
		"localhost",
	}

	certPEM2, _, err := caInstance.GenerateCertificate(serviceName2, serviceIP2, domains2)
	if err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}

	printCertificateDetails("DNS Only", certPEM2)

	// Test case 3: Only IP addresses
	fmt.Println("\n=== Test Case 3: Only IP Addresses ===")
	serviceName3 := "ip-only-service"
	serviceIP3 := "203.0.113.10" // Example IP from RFC 5737
	domains3 := []string{
		"192.168.1.100",
		"10.0.0.1",
		"172.16.0.1",
		"::1",
		"2001:db8::1",
	}

	certPEM3, _, err := caInstance.GenerateCertificate(serviceName3, serviceIP3, domains3)
	if err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}

	printCertificateDetails("IP Only", certPEM3)
}

func printCertificateDetails(testName, certPEM string) {
	// Parse the certificate
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		log.Fatal("Failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse certificate: %v", err)
	}

	fmt.Printf("Test: %s\n", testName)
	fmt.Printf("  CommonName: %s\n", cert.Subject.CommonName)
	fmt.Printf("  DNSNames: %v (count: %d)\n", cert.DNSNames, len(cert.DNSNames))
	fmt.Printf("  IPAddresses: %v (count: %d)\n", cert.IPAddresses, len(cert.IPAddresses))

	// Show the difference in validation
	fmt.Printf("  Certificate validates for:\n")
	for _, dns := range cert.DNSNames {
		fmt.Printf("    - https://%s (DNS validation)\n", dns)
	}
	for _, ip := range cert.IPAddresses {
		fmt.Printf("    - https://%s (IP validation)\n", ip.String())
	}
	fmt.Println()
}
