package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
	// Create a temporary directory for this example
	tempDir, err := os.MkdirTemp("", "ca-threadsafe-example")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Testing thread safety with persistence directory: %s\n\n", tempDir)

	// Create CA with persistence
	config := ca.DefaultCAConfig()
	config.PersistDir = tempDir

	certAuthority, err := ca.NewCA(config)
	if err != nil {
		log.Fatalf("Failed to create CA: %v", err)
	}

	fmt.Println("=== Thread Safety Test ===")
	fmt.Println("Issuing 50 certificates concurrently from 10 goroutines...")

	const numGoroutines = 10
	const certsPerGoroutine = 5
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*certsPerGoroutine)

	start := time.Now()

	// Launch goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < certsPerGoroutine; j++ {
				serviceName := fmt.Sprintf("service-g%d-c%d", goroutineID, j)
				domains := []string{fmt.Sprintf("%s.local", serviceName)}

				req := ca.CertRequest{
					ServiceName: serviceName,
					ServiceIP:   "127.0.0.1",
					Domains:     domains,
				}

				_, err := certAuthority.IssueServiceCertificate(req)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d cert %d failed: %v", goroutineID, j, err)
					return
				}

				// Small delay to increase chance of race conditions
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	duration := time.Since(start)

	// Check for errors
	var errorCount int
	for err := range errors {
		fmt.Printf("❌ Error: %v\n", err)
		errorCount++
	}

	if errorCount > 0 {
		log.Fatalf("Failed with %d errors", errorCount)
	}

	// Verify all certificates were stored
	certs := certAuthority.GetIssuedCertificates()
	expectedCount := numGoroutines * certsPerGoroutine

	fmt.Printf("✅ Successfully issued %d certificates in %v\n", len(certs), duration)
	fmt.Printf("Expected: %d, Actual: %d\n", expectedCount, len(certs))

	if len(certs) == expectedCount {
		fmt.Println("✅ All certificates stored correctly - thread safety verified!")
	} else {
		fmt.Printf("❌ Certificate count mismatch - possible race condition!\n")
		os.Exit(1)
	}

	// Verify count method also works
	count := certAuthority.GetCertificateCount()
	if count == expectedCount {
		fmt.Printf("✅ GetCertificateCount() returned correct value: %d\n", count)
	} else {
		fmt.Printf("❌ GetCertificateCount() returned wrong value: %d (expected %d)\n", count, expectedCount)
		os.Exit(1)
	}

	// Test concurrent reads while writing
	fmt.Println("\n=== Concurrent Read/Write Test ===")
	fmt.Println("Reading certificates while issuing new ones...")

	var readWg sync.WaitGroup
	var writeWg sync.WaitGroup
	readErrors := make(chan error, 10)
	writeErrors := make(chan error, 5)

	// Start readers
	for i := 0; i < 10; i++ {
		readWg.Add(1)
		go func(readerID int) {
			defer readWg.Done()

			for j := 0; j < 20; j++ {
				// Read all certificates
				certs := certAuthority.GetIssuedCertificates()
				if len(certs) < expectedCount {
					readErrors <- fmt.Errorf("reader %d: got %d certs, expected at least %d", readerID, len(certs), expectedCount)
					return
				}

				// Try to get a specific certificate
				if len(certs) > 0 {
					cert, exists := certAuthority.GetCertificateBySerial(certs[0].SerialNumber)
					if !exists || cert == nil {
						readErrors <- fmt.Errorf("reader %d: failed to get certificate by serial", readerID)
						return
					}
				}

				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// Start writers
	for i := 0; i < 5; i++ {
		writeWg.Add(1)
		go func(writerID int) {
			defer writeWg.Done()

			for j := 0; j < 3; j++ {
				serviceName := fmt.Sprintf("concurrent-w%d-c%d", writerID, j)
				req := ca.CertRequest{
					ServiceName: serviceName,
					ServiceIP:   "127.0.0.1",
					Domains:     []string{fmt.Sprintf("%s.local", serviceName)},
				}

				_, err := certAuthority.IssueServiceCertificate(req)
				if err != nil {
					writeErrors <- fmt.Errorf("writer %d cert %d failed: %v", writerID, j, err)
					return
				}

				time.Sleep(2 * time.Millisecond)
			}
		}(i)
	}

	// Wait for all operations
	writeWg.Wait()
	readWg.Wait()
	close(readErrors)
	close(writeErrors)

	// Check for errors
	var totalErrors int
	for err := range readErrors {
		fmt.Printf("❌ Read error: %v\n", err)
		totalErrors++
	}
	for err := range writeErrors {
		fmt.Printf("❌ Write error: %v\n", err)
		totalErrors++
	}

	if totalErrors == 0 {
		fmt.Println("✅ Concurrent read/write operations completed successfully!")
	} else {
		log.Fatalf("Failed with %d errors in concurrent operations", totalErrors)
	}

	// Final verification
	finalCerts := certAuthority.GetIssuedCertificates()
	finalExpected := expectedCount + (5 * 3) // Original + new writer certs
	fmt.Printf("\nFinal certificate count: %d (expected %d)\n", len(finalCerts), finalExpected)

	if len(finalCerts) == finalExpected {
		fmt.Println("🎉 Thread safety and concurrent operations verified successfully!")
	} else {
		fmt.Printf("❌ Final count mismatch - possible data corruption!\n")
		os.Exit(1)
	}
}
