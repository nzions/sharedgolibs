package ca

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"testing"
)

func TestDebugGenerateCertificate(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	serviceName := "test-service"
	serviceIP := "192.168.1.100"
	domains := []string{"test.example.com", "test2.example.com", "127.0.0.1"}

	fmt.Printf("Input: serviceName=%s, serviceIP=%s, domains=%v\n", serviceName, serviceIP, domains)

	// Let's see what net.ParseIP returns for these
	fmt.Printf("net.ParseIP(%q) = %v\n", serviceIP, net.ParseIP(serviceIP))
	for _, domain := range domains {
		fmt.Printf("net.ParseIP(%q) = %v\n", domain, net.ParseIP(domain))
	}

	certPEM, _, err := ca.GenerateCertificate(serviceName, serviceIP, domains)
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// Parse the certificate
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		t.Fatal("Failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	fmt.Printf("Generated CommonName: %s\n", cert.Subject.CommonName)
	fmt.Printf("Generated DNSNames: %v (count: %d)\n", cert.DNSNames, len(cert.DNSNames))
	fmt.Printf("Generated IPAddresses: %v (count: %d)\n", cert.IPAddresses, len(cert.IPAddresses))
}
