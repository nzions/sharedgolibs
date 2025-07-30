// Package main demonstrates the new V2 API for certificate generation
// with automatic IP detection and CN selection.
package main

import (
	"fmt"
	"log"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
	fmt.Println("=== V2 API Certificate Generation Demo ===")

	// Create a new CA
	certAuthority, err := ca.NewCA(nil)
	if err != nil {
		log.Fatalf("Failed to create CA: %v", err)
	}

	// Example 1: Service with mixed SANs (hostnames and IPs)
	fmt.Println("\n1. Generating certificate with mixed SANs...")
	req1 := ca.CertRequestV2{
		ServiceName: "web-service",
		SANs:        []string{"api.example.com", "web.local", "192.168.1.100", "127.0.0.1"},
	}

	resp1, err := certAuthority.IssueServiceCertificateV2(req1)
	if err != nil {
		log.Fatalf("Failed to issue certificate: %v", err)
	}

	fmt.Printf("✅ Certificate generated for: %s\n", req1.ServiceName)
	fmt.Printf("   Certificate length: %d bytes\n", len(resp1.Certificate))
	fmt.Printf("   Private key length: %d bytes\n", len(resp1.PrivateKey))
	fmt.Printf("   Expected CN: api.example.com (first non-IP SAN)\n")

	// Example 2: Service with only hostnames
	fmt.Println("\n2. Generating certificate with only hostnames...")
	req2 := ca.CertRequestV2{
		ServiceName: "auth-service",
		SANs:        []string{"auth.example.com", "authentication.local"},
	}

	resp2, err := certAuthority.IssueServiceCertificateV2(req2)
	if err != nil {
		log.Fatalf("Failed to issue certificate: %v", err)
	}

	fmt.Printf("✅ Certificate generated for: %s\n", req2.ServiceName)
	fmt.Printf("   Certificate length: %d bytes\n", len(resp2.Certificate))
	fmt.Printf("   Expected CN: auth.example.com (first non-IP SAN)\n")

	// Example 3: Service with only IP addresses (fallback to serviceName.local)
	fmt.Println("\n3. Generating certificate with only IP addresses...")
	req3 := ca.CertRequestV2{
		ServiceName: "internal-service",
		SANs:        []string{"10.0.0.100", "172.16.1.50"},
	}

	resp3, err := certAuthority.IssueServiceCertificateV2(req3)
	if err != nil {
		log.Fatalf("Failed to issue certificate: %v", err)
	}

	fmt.Printf("✅ Certificate generated for: %s\n", req3.ServiceName)
	fmt.Printf("   Certificate length: %d bytes\n", len(resp3.Certificate))
	fmt.Printf("   Expected CN: 10.0.0.100 (first IP when no hostnames)\n")

	// Show total certificates issued
	certs := certAuthority.GetIssuedCertificates()
	fmt.Printf("\n📋 Total certificates issued: %d\n", len(certs))

	fmt.Println("\n=== API Comparison ===")
	fmt.Println("Old V1 API: IssueServiceCertificate(serviceName, serviceIP, domains)")
	fmt.Println("New V2 API: IssueServiceCertificateV2(serviceName, sans)")
	fmt.Println("✨ V2 automatically detects IPs vs hostnames and selects appropriate CN!")
	fmt.Println("📋 CN Selection: First hostname → CN, or First IP → CN if no hostnames")
	fmt.Println("🚫 No .local suffix - uses exactly what client provides")

	fmt.Println("\n🎉 V2 API demonstration complete!")
}
