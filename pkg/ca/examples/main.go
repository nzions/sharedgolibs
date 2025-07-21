// Example usage of the SharedGoLibs CA package
// This demonstrates how to create a CA, generate certificates, and run the HTTP server
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [basic|server|custom|client]")
		fmt.Println("Examples:")
		fmt.Println("  go run main.go basic   # Basic CA usage")
		fmt.Println("  go run main.go server  # Start HTTP server")
		fmt.Println("  go run main.go custom  # Custom CA configuration")
		fmt.Println("  go run main.go client  # Generate multiple certificates")
		os.Exit(1)
	}

	example := os.Args[1]

	switch example {
	case "basic":
		basicExample()
	case "server":
		serverExample()
	case "custom":
		customConfigExample()
	case "client":
		multipleServicesExample()
	default:
		fmt.Printf("Unknown example: %s\n", example)
		os.Exit(1)
	}
}

// basicExample demonstrates basic CA usage
func basicExample() {
	fmt.Println("=== Basic CA Example ===")

	// Create a new Certificate Authority with default settings
	certificateAuthority, err := ca.NewCA(nil)
	if err != nil {
		log.Fatalf("Failed to create CA: %v", err)
	}

	fmt.Printf("✅ Created CA: %s\n", certificateAuthority.Certificate().Subject.CommonName)
	fmt.Printf("   Valid until: %s\n", certificateAuthority.Certificate().NotAfter.Format(time.RFC3339))

	// Generate a certificate for a service
	serviceName := "web-service"
	serviceIP := "192.168.1.100"
	domains := []string{"web.example.com", "api.example.com", "localhost"}

	fmt.Printf("\n🔐 Generating certificate for %s...\n", serviceName)
	certPEM, keyPEM, err := certificateAuthority.GenerateCertificate(serviceName, serviceIP, domains)
	if err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}

	fmt.Printf("✅ Certificate generated successfully\n")
	fmt.Printf("   Domains: %v\n", domains)
	fmt.Printf("   Certificate length: %d bytes\n", len(certPEM))
	fmt.Printf("   Private key length: %d bytes\n", len(keyPEM))

	// Save certificates to files
	if err := os.WriteFile("service.crt", []byte(certPEM), 0644); err != nil {
		log.Printf("Failed to save certificate: %v", err)
	} else {
		fmt.Println("   Saved to: service.crt")
	}

	if err := os.WriteFile("service.key", []byte(keyPEM), 0600); err != nil {
		log.Printf("Failed to save private key: %v", err)
	} else {
		fmt.Println("   Saved to: service.key")
	}

	// Save CA certificate
	caCertPEM := certificateAuthority.CertificatePEM()
	if err := os.WriteFile("ca.crt", caCertPEM, 0644); err != nil {
		log.Printf("Failed to save CA certificate: %v", err)
	} else {
		fmt.Println("   CA saved to: ca.crt")
	}

	// Show certificate store info
	fmt.Printf("\n📋 Certificate Store:\n")
	fmt.Printf("   Issued certificates: %d\n", certificateAuthority.GetCertificateCount())

	certs := certificateAuthority.GetIssuedCertificates()
	for _, cert := range certs {
		fmt.Printf("   - %s (Serial: %s, Expires: %s)\n",
			cert.ServiceName,
			cert.SerialNumber,
			cert.ExpiresAt.Format("2006-01-02"))
	}
}

// serverExample demonstrates running the HTTP server
func serverExample() {
	fmt.Println("=== CA Server Example ===")

	// Create server with default configuration
	server, err := ca.NewServer(nil)
	if err != nil {
		log.Fatalf("Failed to create CA server: %v", err)
	}

	fmt.Println("🚀 Starting CA server...")
	fmt.Println("   Port: 8090")
	fmt.Println("   Endpoints:")
	fmt.Println("     GET  /ca       - Download CA certificate")
	fmt.Println("     POST /cert     - Request service certificate")
	fmt.Println("     GET  /health   - Health check")
	fmt.Println("     GET  /ui/      - Web UI dashboard")
	fmt.Println("")
	fmt.Println("🌐 Open http://localhost:8090/ui/ in your browser")
	fmt.Println("")
	fmt.Println("📝 Example API request:")
	fmt.Println(`   curl -X POST http://localhost:8090/cert \
     -H "Content-Type: application/json" \
     -d '{
       "service_name": "my-service",
       "service_ip": "192.168.1.10",
       "domains": ["my-service.local", "api.my-service.local"]
     }'`)
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop the server")

	// Start the server (this blocks)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// customConfigExample demonstrates custom CA configuration
func customConfigExample() {
	fmt.Println("=== Custom CA Configuration Example ===")

	// Create custom CA configuration
	config := &ca.CAConfig{
		Country:            []string{"CA"},
		Province:           []string{"Ontario"},
		Locality:           []string{"Toronto"},
		Organization:       []string{"Example Corp"},
		OrganizationalUnit: []string{"IT Department"},
		CommonName:         "Example Corp Root CA",
		ValidityPeriod:     2 * 365 * 24 * time.Hour, // 2 years
		KeySize:            4096,                     // Larger key size
	}

	fmt.Printf("🏗️  Creating CA with custom configuration:\n")
	fmt.Printf("   Organization: %s\n", config.Organization[0])
	fmt.Printf("   Common Name: %s\n", config.CommonName)
	fmt.Printf("   Validity Period: %v\n", config.ValidityPeriod)
	fmt.Printf("   Key Size: %d bits\n", config.KeySize)

	certificateAuthority, err := ca.NewCA(config)
	if err != nil {
		log.Fatalf("Failed to create CA: %v", err)
	}

	fmt.Printf("✅ CA created successfully\n")

	// Get CA info
	caInfo := certificateAuthority.GetCAInfo()
	fmt.Printf("\n📄 CA Information:\n")
	for key, value := range caInfo {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// Generate a certificate with the custom CA
	fmt.Printf("\n🔐 Generating certificate with custom CA...\n")
	certPEM, _, err := certificateAuthority.GenerateCertificate(
		"secure-service",
		"10.0.0.50",
		[]string{"secure.example.com", "internal.example.com"},
	)
	if err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}

	fmt.Printf("✅ Certificate generated with custom CA\n")
	fmt.Printf("   Certificate contains: %d characters\n", len(certPEM))

	// Save the custom CA certificate
	caCertPEM := certificateAuthority.CertificatePEM()
	if err := os.WriteFile("custom-ca.crt", caCertPEM, 0644); err != nil {
		log.Printf("Failed to save CA certificate: %v", err)
	} else {
		fmt.Println("   Custom CA saved to: custom-ca.crt")
	}
}

// multipleServicesExample demonstrates generating certificates for multiple services
func multipleServicesExample() {
	fmt.Println("=== Multiple Services Example ===")

	// Create CA
	certificateAuthority, err := ca.NewCA(nil)
	if err != nil {
		log.Fatalf("Failed to create CA: %v", err)
	}

	// Define multiple services
	services := []struct {
		name    string
		ip      string
		domains []string
	}{
		{
			name:    "web-frontend",
			ip:      "10.0.1.10",
			domains: []string{"web.example.com", "www.example.com", "frontend.local"},
		},
		{
			name:    "api-backend",
			ip:      "10.0.1.20",
			domains: []string{"api.example.com", "v1.api.example.com", "backend.local"},
		},
		{
			name:    "database",
			ip:      "10.0.1.30",
			domains: []string{"db.internal", "postgres.internal"},
		},
		{
			name:    "redis-cache",
			ip:      "10.0.1.40",
			domains: []string{"redis.internal", "cache.internal"},
		},
		{
			name:    "monitoring",
			ip:      "10.0.1.50",
			domains: []string{"monitor.example.com", "grafana.internal", "prometheus.internal"},
		},
	}

	fmt.Printf("🏭 Generating certificates for %d services...\n", len(services))

	// Generate certificates for all services
	for i, service := range services {
		fmt.Printf("\n[%d/%d] Generating certificate for %s\n", i+1, len(services), service.name)
		fmt.Printf("        IP: %s\n", service.ip)
		fmt.Printf("        Domains: %v\n", service.domains)

		// Generate certificate
		certPEM, keyPEM, err := certificateAuthority.GenerateCertificate(service.name, service.ip, service.domains)
		if err != nil {
			log.Printf("❌ Failed to generate certificate for %s: %v", service.name, err)
			continue
		}

		// Save certificate files
		certFile := fmt.Sprintf("%s.crt", service.name)
		keyFile := fmt.Sprintf("%s.key", service.name)

		if err := os.WriteFile(certFile, []byte(certPEM), 0644); err != nil {
			log.Printf("Failed to save certificate: %v", err)
		}

		if err := os.WriteFile(keyFile, []byte(keyPEM), 0600); err != nil {
			log.Printf("Failed to save private key: %v", err)
		}

		fmt.Printf("        ✅ Saved: %s, %s\n", certFile, keyFile)
	}

	// Show summary
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   Total certificates generated: %d\n", certificateAuthority.GetCertificateCount())

	// Show all issued certificates
	fmt.Printf("\n📋 Issued Certificates:\n")
	certs := certificateAuthority.GetIssuedCertificates()
	for _, cert := range certs {
		fmt.Printf("   🔐 %s\n", cert.ServiceName)
		fmt.Printf("      Serial: %s\n", cert.SerialNumber)
		fmt.Printf("      Domains: %v\n", cert.Domains)
		fmt.Printf("      Issued: %s\n", cert.IssuedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("      Expires: %s\n", cert.ExpiresAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("      Valid for: %.0f days\n", time.Until(cert.ExpiresAt).Hours()/24)
		fmt.Println()
	}

	// Save CA certificate
	caCertPEM := certificateAuthority.CertificatePEM()
	if err := os.WriteFile("services-ca.crt", caCertPEM, 0644); err != nil {
		log.Printf("Failed to save CA certificate: %v", err)
	} else {
		fmt.Printf("📄 CA certificate saved to: services-ca.crt\n")
	}

	// Show example cert request format
	fmt.Printf("\n📝 Example certificate request format:\n")
	exampleReq := ca.CertRequest{
		ServiceName: "new-service",
		ServiceIP:   "10.0.1.60",
		Domains:     []string{"new-service.example.com", "ns.internal"},
	}

	reqJSON, _ := json.MarshalIndent(exampleReq, "", "  ")
	fmt.Printf("%s\n", string(reqJSON))

	fmt.Printf("\n🔧 Files created:\n")
	for _, service := range services {
		fmt.Printf("   %s.crt, %s.key\n", service.name, service.name)
	}
	fmt.Printf("   services-ca.crt\n")

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   1. Distribute services-ca.crt to all clients\n")
	fmt.Printf("   2. Configure each service to use its certificate and key\n")
	fmt.Printf("   3. Set up automatic certificate rotation (certificates expire in 30 days)\n")
}
