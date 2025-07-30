package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CertStorage defines the interface for certificate storage
type CertStorage interface {
	// GenerateAndStore generates a certificate and stores it atomically
	GenerateAndStore(ca *CA, serviceName, serviceIP string, domains []string) (string, string, error)
	// GenerateAndStoreV2 generates a certificate using V2 API with automatic IP detection
	GenerateAndStoreV2(ca *CA, serviceName string, sans []string) (string, string, error)
	// GetAll returns all stored certificates
	GetAll() ([]*IssuedCert, error)
	// GetBySerial returns a certificate by serial number
	GetBySerial(serial string) (*IssuedCert, error)
	// Count returns the number of stored certificates
	Count() (int, error)
}

// RAMStorage implements in-memory certificate storage
type RAMStorage struct {
	certs map[string]*IssuedCert
	mutex sync.RWMutex
}

// NewRAMStorage creates a new in-memory certificate storage for issued certificates.
// Returns a pointer to a RAMStorage instance.
func NewRAMStorage() *RAMStorage {
	return &RAMStorage{
		certs: make(map[string]*IssuedCert),
	}
}

// GenerateAndStore generates a certificate and stores it atomically in memory.
// Returns PEM-encoded certificate, private key, and error if any.
func (s *RAMStorage) GenerateAndStore(ca *CA, serviceName, serviceIP string, domains []string) (string, string, error) {
	// Generate the certificate
	serviceCertPEM, serviceKeyPEM, issuedCert, err := s.generateCertificate(ca, serviceName, serviceIP, domains)
	if err != nil {
		return "", "", err
	}

	// Store atomically
	s.mutex.Lock()
	s.certs[issuedCert.SerialNumber] = issuedCert
	s.mutex.Unlock()

	return serviceCertPEM, serviceKeyPEM, nil
}

// GenerateAndStoreV2 generates a certificate using the V2 API with automatic IP detection
// and stores it atomically in memory.
func (s *RAMStorage) GenerateAndStoreV2(ca *CA, serviceName string, sans []string) (string, string, error) {
	// Use the existing generateCertificate method but pass empty serviceIP since
	// IP addresses are now included in the sans array
	serviceCertPEM, serviceKeyPEM, issuedCert, err := s.generateCertificate(ca, serviceName, "", sans)
	if err != nil {
		return "", "", err
	}

	// Store atomically
	s.mutex.Lock()
	s.certs[issuedCert.SerialNumber] = issuedCert
	s.mutex.Unlock()

	return serviceCertPEM, serviceKeyPEM, nil
}

// GetAll returns all certificates from memory.
// Returns a slice of IssuedCert pointers and error if retrieval fails.
func (s *RAMStorage) GetAll() ([]*IssuedCert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	certs := make([]*IssuedCert, 0, len(s.certs))
	for _, cert := range s.certs {
		certs = append(certs, cert)
	}
	return certs, nil
}

// GetBySerial returns a certificate by serial number from memory.
// Returns the certificate and error if not found.
func (s *RAMStorage) GetBySerial(serial string) (*IssuedCert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cert, exists := s.certs[serial]
	if !exists {
		return nil, nil
	}
	return cert, nil
}

// Count returns the number of certificates in memory.
// Returns the count and error if retrieval fails.
func (s *RAMStorage) Count() (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.certs), nil
}

// DiskStorage implements persistent certificate storage
type DiskStorage struct {
	persistDir string
	certs      map[string]*IssuedCert
	mutex      sync.RWMutex
}

// NewDiskStorage creates a new disk-based certificate storage for issued certificates.
// Returns a pointer to DiskStorage and error if initialization fails.
func NewDiskStorage(persistDir string) (*DiskStorage, error) {
	storage := &DiskStorage{
		persistDir: persistDir,
		certs:      make(map[string]*IssuedCert),
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(persistDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create persist directory %s: %w", persistDir, err)
	}

	// Check if directory is writable
	testFile := filepath.Join(persistDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return nil, fmt.Errorf("persist directory %s is not writable: %w", persistDir, err)
	}
	os.Remove(testFile) // Clean up test file

	// Load existing certificates
	if err := storage.loadFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load certificates from disk: %w", err)
	}

	return storage, nil
}

// GenerateAndStore generates a certificate and stores it atomically to disk.
// Returns PEM-encoded certificate, private key, and error if any.
func (s *DiskStorage) GenerateAndStore(ca *CA, serviceName, serviceIP string, domains []string) (string, string, error) {
	// Generate the certificate
	serviceCertPEM, serviceKeyPEM, issuedCert, err := s.generateCertificate(ca, serviceName, serviceIP, domains)
	if err != nil {
		return "", "", err
	}

	// Store atomically (both in memory and on disk)
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.certs[issuedCert.SerialNumber] = issuedCert
	err = s.saveToDisk()
	if err != nil {
		// Rollback the in-memory change if disk save fails
		delete(s.certs, issuedCert.SerialNumber)
		return "", "", fmt.Errorf("failed to persist certificate to disk: %w", err)
	}

	return serviceCertPEM, serviceKeyPEM, nil
}

// GenerateAndStoreV2 generates a certificate using the V2 API with automatic IP detection
// and stores it atomically to disk.
func (s *DiskStorage) GenerateAndStoreV2(ca *CA, serviceName string, sans []string) (string, string, error) {
	// Use the existing generateCertificate method but pass empty serviceIP since
	// IP addresses are now included in the sans array
	serviceCertPEM, serviceKeyPEM, issuedCert, err := s.generateCertificate(ca, serviceName, "", sans)
	if err != nil {
		return "", "", err
	}

	// Store atomically (both in memory and on disk)
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.certs[issuedCert.SerialNumber] = issuedCert
	err = s.saveToDisk()
	if err != nil {
		// Rollback the in-memory change if disk save fails
		delete(s.certs, issuedCert.SerialNumber)
		return "", "", fmt.Errorf("failed to persist certificate to disk: %w", err)
	}

	return serviceCertPEM, serviceKeyPEM, nil
}

// GetAll returns all certificates from disk storage.
// Returns a slice of IssuedCert pointers and error if retrieval fails.
func (s *DiskStorage) GetAll() ([]*IssuedCert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	certs := make([]*IssuedCert, 0, len(s.certs))
	for _, cert := range s.certs {
		certs = append(certs, cert)
	}
	return certs, nil
}

// GetBySerial returns a certificate by serial number from disk storage.
// Returns the certificate and error if not found.
func (s *DiskStorage) GetBySerial(serial string) (*IssuedCert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cert, exists := s.certs[serial]
	if !exists {
		return nil, nil
	}
	return cert, nil
}

// Count returns the number of certificates in disk storage.
// Returns the count and error if retrieval fails.
func (s *DiskStorage) Count() (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.certs), nil
}

// generateCertificate creates a new certificate for the given service and domains (RAMStorage).
// Returns PEM-encoded certificate, private key, IssuedCert, and error if any.
func (s *RAMStorage) generateCertificate(ca *CA, serviceName, serviceIP string, domains []string) (string, string, *IssuedCert, error) {
	return generateCertificateInternal(ca, serviceName, serviceIP, domains)
}

// generateCertificate creates a new certificate for the given service and domains (DiskStorage).
// Returns PEM-encoded certificate, private key, IssuedCert, and error if any.
func (s *DiskStorage) generateCertificate(ca *CA, serviceName, serviceIP string, domains []string) (string, string, *IssuedCert, error) {
	return generateCertificateInternal(ca, serviceName, serviceIP, domains)
}

// generateCertificateInternal contains the shared certificate generation logic for both storage types.
// Returns PEM-encoded certificate, private key, IssuedCert, and error if any.
func generateCertificateInternal(ca *CA, serviceName, serviceIP string, domains []string) (string, string, *IssuedCert, error) {
	// Generate service private key
	serviceKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate service private key: %w", err)
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1000000000))
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Validate that domains/SANs is not empty
	if len(domains) == 0 {
		return "", "", nil, fmt.Errorf("domains/SANs cannot be empty")
	}

	// Determine the CommonName with new logic:
	// 1. Use first non-IP domain if available
	// 2. If only IPs, use first IP as CN
	// 3. Never add .local suffix - use exactly what client supplies
	var commonName string
	var foundHostname bool

	// First pass: look for non-IP domains
	for _, domain := range domains {
		if net.ParseIP(domain) == nil {
			// This is not an IP address, use it as CN
			commonName = domain
			foundHostname = true
			break
		}
	}

	// If no hostname found, use first IP as CN
	if !foundHostname {
		commonName = domains[0] // Use first IP address as CN
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"SharedGoLibs Services"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Keep track of added IPs to avoid duplicates
	addedIPs := make(map[string]bool)

	// Add service IP first if provided
	if serviceIP != "" && serviceIP != "0.0.0.0" {
		if ip := net.ParseIP(serviceIP); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
			addedIPs[serviceIP] = true
		}
	}

	// Process domains - separate DNS names from IP addresses
	for _, domain := range domains {
		if ip := net.ParseIP(domain); ip != nil {
			// This is an IP address - add only if not already present
			if !addedIPs[domain] {
				template.IPAddresses = append(template.IPAddresses, ip)
				addedIPs[domain] = true
			}
		} else {
			// This is a DNS name
			template.DNSNames = append(template.DNSNames, domain)
		}
	}

	// Generate certificate using CA (need to protect CA access)
	ca.mutex.RLock()
	caCert := ca.cert
	caKey := ca.privateKey
	ca.mutex.RUnlock()

	if caCert == nil || caKey == nil {
		return "", "", nil, fmt.Errorf("CA not properly initialized")
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &serviceKey.PublicKey, caKey)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate as PEM
	serviceCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key as PEM
	serviceKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serviceKey),
	})

	// Create IssuedCert record
	issuedCert := &IssuedCert{
		ServiceName:  serviceName,
		Domains:      domains,
		IssuedAt:     time.Now(),
		ExpiresAt:    template.NotAfter,
		Certificate:  string(serviceCertPEM),
		PrivateKey:   string(serviceKeyPEM),
		SerialNumber: fmt.Sprintf("%x", serialNumber),
	}

	return string(serviceCertPEM), string(serviceKeyPEM), issuedCert, nil
}

// saveToDisk saves the certificate store to disk (must be called with mutex locked).
// Returns error if saving fails.
func (s *DiskStorage) saveToDisk() error {
	certStorePath := filepath.Join(s.persistDir, "cert-store.json")

	data, err := json.MarshalIndent(s.certs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal certificate store: %w", err)
	}

	if err := os.WriteFile(certStorePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save certificate store: %w", err)
	}

	return nil
}

// loadFromDisk loads the certificate store from disk.
// Returns error if loading fails.
func (s *DiskStorage) loadFromDisk() error {
	certStorePath := filepath.Join(s.persistDir, "cert-store.json")

	// Check if file exists
	if _, err := os.Stat(certStorePath); os.IsNotExist(err) {
		return nil // No existing store to load
	}

	data, err := os.ReadFile(certStorePath)
	if err != nil {
		return fmt.Errorf("failed to read certificate store: %w", err)
	}

	if len(data) == 0 {
		return nil // Empty file, nothing to load
	}

	if err := json.Unmarshal(data, &s.certs); err != nil {
		return fmt.Errorf("failed to unmarshal certificate store: %w", err)
	}

	return nil
}
