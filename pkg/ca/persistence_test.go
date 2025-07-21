package ca

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistence(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Test disk storage
	t.Run("DiskStorage", func(t *testing.T) {
		// Create CA with persistence
		config := DefaultCAConfig()
		config.PersistDir = tempDir

		ca1, err := NewCA(config)
		if err != nil {
			t.Fatalf("Failed to create CA with persistence: %v", err)
		}

		// Issue a certificate
		req := CertRequest{
			ServiceName: "test-service",
			ServiceIP:   "127.0.0.1",
			Domains:     []string{"test.local"},
		}

		resp1, err := ca1.IssueServiceCertificate(req)
		if err != nil {
			t.Fatalf("Failed to issue certificate: %v", err)
		}

		// Verify certificate was stored
		certs1 := ca1.GetIssuedCertificates()
		if len(certs1) != 1 {
			t.Fatalf("Expected 1 certificate, got %d", len(certs1))
		}

		// Create a new CA instance with the same persist dir
		ca2, err := NewCA(config)
		if err != nil {
			t.Fatalf("Failed to create second CA instance: %v", err)
		}

		// Verify the CA certificate is the same
		if string(ca1.CertificatePEM()) != string(ca2.CertificatePEM()) {
			t.Fatalf("CA certificates don't match after persistence")
		}

		// Verify the issued certificate was loaded
		certs2 := ca2.GetIssuedCertificates()
		if len(certs2) != 1 {
			t.Fatalf("Expected 1 certificate after reload, got %d", len(certs2))
		}

		if certs1[0].SerialNumber != certs2[0].SerialNumber {
			t.Fatalf("Certificate serial numbers don't match after persistence")
		}

		// Issue another certificate with the second CA
		req2 := CertRequest{
			ServiceName: "test-service-2",
			ServiceIP:   "127.0.0.1",
			Domains:     []string{"test2.local"},
		}

		resp2, err := ca2.IssueServiceCertificate(req2)
		if err != nil {
			t.Fatalf("Failed to issue second certificate: %v", err)
		}

		// Verify both certificates are stored
		certs3 := ca2.GetIssuedCertificates()
		if len(certs3) != 2 {
			t.Fatalf("Expected 2 certificates, got %d", len(certs3))
		}

		// Verify responses are different
		if resp1.Certificate == resp2.Certificate {
			t.Fatalf("Certificate responses should be different")
		}
	})

	// Test RAM storage
	t.Run("RAMStorage", func(t *testing.T) {
		// Create CA without persistence
		config := DefaultCAConfig()
		config.PersistDir = ""

		ca1, err := NewCA(config)
		if err != nil {
			t.Fatalf("Failed to create CA without persistence: %v", err)
		}

		// Issue a certificate
		req := CertRequest{
			ServiceName: "test-service",
			ServiceIP:   "127.0.0.1",
			Domains:     []string{"test.local"},
		}

		_, err = ca1.IssueServiceCertificate(req)
		if err != nil {
			t.Fatalf("Failed to issue certificate: %v", err)
		}

		// Verify certificate was stored
		certs1 := ca1.GetIssuedCertificates()
		if len(certs1) != 1 {
			t.Fatalf("Expected 1 certificate, got %d", len(certs1))
		}

		// Create a new CA instance (should be completely fresh)
		ca2, err := NewCA(config)
		if err != nil {
			t.Fatalf("Failed to create second CA instance: %v", err)
		}

		// Verify the CA certificate is different (new CA created)
		if string(ca1.CertificatePEM()) == string(ca2.CertificatePEM()) {
			t.Fatalf("CA certificates should be different without persistence")
		}

		// Verify no certificates are loaded
		certs2 := ca2.GetIssuedCertificates()
		if len(certs2) != 0 {
			t.Fatalf("Expected 0 certificates in new CA, got %d", len(certs2))
		}
	})

	// Test storage operations are thread-safe
	t.Run("ThreadSafety", func(t *testing.T) {
		config := DefaultCAConfig()
		config.PersistDir = filepath.Join(tempDir, "threadsafe")

		ca, err := NewCA(config)
		if err != nil {
			t.Fatalf("Failed to create CA: %v", err)
		}

		// Issue certificates concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				req := CertRequest{
					ServiceName: fmt.Sprintf("test-service-%d", id), // Unique service name
					ServiceIP:   "127.0.0.1",
					Domains:     []string{fmt.Sprintf("test%d.local", id)}, // Unique domain
				}

				_, err := ca.IssueServiceCertificate(req)
				if err != nil {
					t.Errorf("Failed to issue certificate %d: %v", id, err)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			select {
			case <-done:
				// OK
			case <-time.After(5 * time.Second):
				t.Fatalf("Timeout waiting for concurrent certificate issuance")
			}
		}

		// Verify all certificates were stored
		certs := ca.GetIssuedCertificates()
		if len(certs) != 10 {
			t.Fatalf("Expected 10 certificates, got %d", len(certs))
		}

		// Verify count is correct
		count := ca.GetCertificateCount()
		if count != 10 {
			t.Fatalf("Expected count 10, got %d", count)
		}
	})

	// Test disk storage with unwritable directory
	t.Run("UnwritableDirectory", func(t *testing.T) {
		// Create a directory and make it unwritable (if running as non-root)
		unwritableDir := filepath.Join(tempDir, "unwritable")
		err := os.MkdirAll(unwritableDir, 0555) // Read and execute only
		if err != nil {
			t.Fatalf("Failed to create unwritable directory: %v", err)
		}

		config := DefaultCAConfig()
		config.PersistDir = unwritableDir

		_, err = NewCA(config)
		if err == nil {
			t.Fatalf("Expected error when creating CA with unwritable directory")
		}
	})
}

func TestStorageInterface(t *testing.T) {
	// Create a CA for testing storage operations
	ca, err := NewCA(DefaultCAConfig())
	if err != nil {
		t.Fatalf("Failed to create CA for testing: %v", err)
	}

	t.Run("RAMStorage", func(t *testing.T) {
		storage := NewRAMStorage()

		// Test GenerateAndStore
		certPEM, keyPEM, err := storage.GenerateAndStore(ca, "test-service", "127.0.0.1", []string{"test.local"})
		if err != nil {
			t.Fatalf("Failed to generate and store certificate: %v", err)
		}
		if certPEM == "" || keyPEM == "" {
			t.Fatalf("Expected non-empty certificate and key")
		}

		// Test GetAll
		certs, err := storage.GetAll()
		if err != nil {
			t.Fatalf("Failed to get all certificates: %v", err)
		}
		if len(certs) != 1 {
			t.Fatalf("Expected 1 certificate, got %d", len(certs))
		}

		// Test GetBySerial
		retrievedCert, err := storage.GetBySerial(certs[0].SerialNumber)
		if err != nil {
			t.Fatalf("Failed to get certificate by serial: %v", err)
		}
		if retrievedCert == nil {
			t.Fatalf("Expected certificate, got nil")
		}
		if retrievedCert.ServiceName != "test-service" {
			t.Fatalf("Expected service name 'test-service', got %s", retrievedCert.ServiceName)
		}

		// Test Count
		count, err := storage.Count()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count != 1 {
			t.Fatalf("Expected count 1, got %d", count)
		}

		// Test GetBySerial with non-existent serial
		nonExistent, err := storage.GetBySerial("999999")
		if err != nil {
			t.Fatalf("Failed to get non-existent certificate: %v", err)
		}
		if nonExistent != nil {
			t.Fatalf("Expected nil for non-existent certificate, got %v", nonExistent)
		}
	})

	t.Run("DiskStorage", func(t *testing.T) {
		tempDir := t.TempDir()
		storage, err := NewDiskStorage(tempDir)
		if err != nil {
			t.Fatalf("Failed to create disk storage: %v", err)
		}

		// Test GenerateAndStore
		certPEM, keyPEM, err := storage.GenerateAndStore(ca, "test-service", "127.0.0.1", []string{"test.local"})
		if err != nil {
			t.Fatalf("Failed to generate and store certificate: %v", err)
		}
		if certPEM == "" || keyPEM == "" {
			t.Fatalf("Expected non-empty certificate and key")
		}

		// Verify file was created
		certFile := filepath.Join(tempDir, "cert-store.json")
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			t.Fatalf("Certificate store file was not created")
		}

		// Test GetAll
		certs, err := storage.GetAll()
		if err != nil {
			t.Fatalf("Failed to get all certificates: %v", err)
		}
		if len(certs) != 1 {
			t.Fatalf("Expected 1 certificate, got %d", len(certs))
		}

		// Test GetBySerial
		retrievedCert, err := storage.GetBySerial(certs[0].SerialNumber)
		if err != nil {
			t.Fatalf("Failed to get certificate by serial: %v", err)
		}
		if retrievedCert == nil {
			t.Fatalf("Expected certificate, got nil")
		}

		// Test Count
		count, err := storage.Count()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count != 1 {
			t.Fatalf("Expected count 1, got %d", count)
		}

		// Create new storage instance with same directory
		storage2, err := NewDiskStorage(tempDir)
		if err != nil {
			t.Fatalf("Failed to create second disk storage instance: %v", err)
		}

		// Verify data was loaded
		certs2, err := storage2.GetAll()
		if err != nil {
			t.Fatalf("Failed to get all certificates from second instance: %v", err)
		}
		if len(certs2) != 1 {
			t.Fatalf("Expected 1 certificate in second instance, got %d", len(certs2))
		}
	})
}
