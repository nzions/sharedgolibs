package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
	// Create a temporary directory for this example
	tempDir, err := os.MkdirTemp("", "ca-persistence-example")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Using persistence directory: %s\n\n", tempDir)

	// Test 1: Create CA with persistence enabled
	fmt.Println("=== Test 1: Creating CA with persistence ===")
	config1 := ca.DefaultCAConfig()
	config1.PersistDir = tempDir

	ca1, err := ca.NewCA(config1)
	if err != nil {
		log.Fatalf("Failed to create CA: %v", err)
	}

	// Issue some certificates
	fmt.Println("Issuing certificates...")
	for i := 1; i <= 3; i++ {
		req := ca.CertRequest{
			ServiceName: fmt.Sprintf("service-%d", i),
			ServiceIP:   "127.0.0.1",
			Domains:     []string{fmt.Sprintf("service-%d.local", i)},
		}

		resp, err := ca1.IssueServiceCertificate(req)
		if err != nil {
			log.Fatalf("Failed to issue certificate %d: %v", i, err)
		}
		fmt.Printf("✅ Issued certificate for %s\n", req.ServiceName)
		_ = resp // Response contains cert, key, and CA cert
	}

	// Check certificate count
	certs1 := ca1.GetIssuedCertificates()
	fmt.Printf("CA instance 1 has %d certificates\n\n", len(certs1))

	// Test 2: Create new CA instance with same persistence directory
	fmt.Println("=== Test 2: Loading existing CA from persistence ===")
	config2 := ca.DefaultCAConfig()
	config2.PersistDir = tempDir

	ca2, err := ca.NewCA(config2)
	if err != nil {
		log.Fatalf("Failed to create second CA: %v", err)
	}

	// Check if certificates were loaded
	certs2 := ca2.GetIssuedCertificates()
	fmt.Printf("CA instance 2 has %d certificates (should be same as instance 1)\n", len(certs2))

	// Verify CA certificates are the same
	if string(ca1.CertificatePEM()) == string(ca2.CertificatePEM()) {
		fmt.Println("✅ CA certificates match - persistence working correctly!")
	} else {
		fmt.Println("❌ CA certificates don't match - persistence failed!")
	}

	// Issue another certificate with the second CA instance
	fmt.Println("\nIssuing another certificate with second CA instance...")
	req := ca.CertRequest{
		ServiceName: "service-4",
		ServiceIP:   "127.0.0.1",
		Domains:     []string{"service-4.local"},
	}

	_, err = ca2.IssueServiceCertificate(req)
	if err != nil {
		log.Fatalf("Failed to issue certificate with second CA: %v", err)
	}
	fmt.Printf("✅ Issued certificate for %s\n", req.ServiceName)

	// Check final count
	certs3 := ca2.GetIssuedCertificates()
	fmt.Printf("CA instance 2 now has %d certificates\n\n", len(certs3))

	// Test 3: Compare with RAM-only CA
	fmt.Println("=== Test 3: RAM-only CA (no persistence) ===")
	config3 := ca.DefaultCAConfig()
	config3.PersistDir = "" // Empty = RAM only

	ca3, err := ca.NewCA(config3)
	if err != nil {
		log.Fatalf("Failed to create RAM-only CA: %v", err)
	}

	// Check certificate count (should be 0)
	certs4 := ca3.GetIssuedCertificates()
	fmt.Printf("RAM-only CA has %d certificates (should be 0)\n", len(certs4))

	// Verify CA certificate is different
	if string(ca2.CertificatePEM()) != string(ca3.CertificatePEM()) {
		fmt.Println("✅ RAM-only CA has different certificate - correct behavior!")
	} else {
		fmt.Println("❌ RAM-only CA has same certificate - this shouldn't happen!")
	}

	// Show what files were created
	fmt.Println("\n=== Files created for persistence ===")
	files, err := os.ReadDir(tempDir)
	if err != nil {
		log.Fatalf("Failed to read persistence directory: %v", err)
	}

	for _, file := range files {
		filePath := filepath.Join(tempDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}
		fmt.Printf("📁 %s (%d bytes, modified: %s)\n",
			file.Name(),
			info.Size(),
			info.ModTime().Format(time.RFC3339))
	}

	fmt.Println("\n🎉 Persistence example completed successfully!")
}
