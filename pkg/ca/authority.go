// Package ca provides Certificate Authority functionality
// Provides dynamic certificate issuance for services and applications
// Acts like Let's Encrypt for development and testing environments
// Version: 1.0.0
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
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	Version = "v1.1.0"
)

// CA represents a Certificate Authority with the ability to issue certificates
type CA struct {
	cert       *x509.Certificate
	privateKey *rsa.PrivateKey
	storage    CertStorage
	mutex      sync.RWMutex // Protects CA certificate and private key
	persistDir string       // Directory for CA persistence (empty = RAM only)
}

// IssuedCert represents a certificate that has been issued by the CA
type IssuedCert struct {
	ServiceName  string    `json:"service_name"`
	Domains      []string  `json:"domains"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Certificate  string    `json:"certificate"`
	PrivateKey   string    `json:"private_key,omitempty"` // Optional for security
	SerialNumber string    `json:"serial_number"`
}

// CertRequest represents a request for a new certificate
type CertRequest struct {
	ServiceName string   `json:"service_name"`
	ServiceIP   string   `json:"service_ip"`
	Domains     []string `json:"domains"`
}

// CertResponse represents the response containing the issued certificate
type CertResponse struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
	CACert      string `json:"ca_cert"`
}

// CAConfig holds configuration options for creating a new CA
type CAConfig struct {
	// Organization details for the CA certificate
	Country            []string
	Province           []string
	Locality           []string
	Organization       []string
	OrganizationalUnit []string
	CommonName         string

	// Certificate validity period
	ValidityPeriod time.Duration

	// Key size for CA private key (default: 4096)
	KeySize int

	// Directory to persist CA data (empty = RAM only)
	PersistDir string
}

// HTTPTransportSettings configures the global HTTP transport
type HTTPTransportSettings struct {
	Timeout               time.Duration
	KeepAlive             time.Duration
	MaxIdleConns          int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
}

// DefaultCAConfig returns sensible defaults for CA configuration
func DefaultCAConfig() *CAConfig {
	return &CAConfig{
		Country:            []string{"US"},
		Province:           []string{"Local"},
		Locality:           []string{"Local"},
		Organization:       []string{"SharedGoLibs Development"},
		OrganizationalUnit: []string{"CA"},
		CommonName:         "SharedGoLibs Root CA",
		ValidityPeriod:     365 * 24 * time.Hour, // 1 year
		KeySize:            4096,
		PersistDir:         "", // RAM only by default
	}
}

// NewCA creates and initializes a new Certificate Authority
func NewCA(config *CAConfig) (*CA, error) {
	if config == nil {
		config = DefaultCAConfig()
	}

	ca := &CA{
		persistDir: config.PersistDir,
	}

	// Initialize storage based on configuration
	var err error
	if config.PersistDir != "" {
		ca.storage, err = NewDiskStorage(config.PersistDir)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize disk storage: %w", err)
		}
		fmt.Printf("[ca] Using disk storage: %s\n", config.PersistDir)
	} else {
		ca.storage = NewRAMStorage()
		fmt.Printf("[ca] Using RAM-only storage\n")
	}

	if err := ca.initialize(config); err != nil {
		return nil, fmt.Errorf("failed to initialize CA: %w", err)
	}

	return ca, nil
}

// initialize sets up the CA certificate and private key
func (ca *CA) initialize(config *CAConfig) error {
	// Try to load existing CA from disk if persistence is enabled
	if ca.persistDir != "" {
		if err := ca.loadCAFromDisk(); err != nil {
			return fmt.Errorf("failed to load CA from disk: %w", err)
		}
	}

	// If we loaded an existing CA, we're done
	if ca.cert != nil && ca.privateKey != nil {
		fmt.Printf("[ca] Loaded existing CA from disk\n")
		return nil
	}

	// Generate new CA private key
	var err error
	ca.privateKey, err = rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %w", err)
	}

	// Create CA certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:            config.Country,
			Province:           config.Province,
			Locality:           config.Locality,
			Organization:       config.Organization,
			OrganizationalUnit: config.OrganizationalUnit,
			CommonName:         config.CommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(config.ValidityPeriod),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
	}

	// Create CA certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &ca.privateKey.PublicKey, ca.privateKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Parse CA certificate
	ca.cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Save the newly created CA to disk
	if err := ca.saveCAKeyToDisk(); err != nil {
		return fmt.Errorf("failed to save CA to disk: %w", err)
	}

	if ca.persistDir != "" {
		fmt.Printf("[ca] Created and saved new CA to disk\n")
	} else {
		fmt.Printf("[ca] Created new CA (RAM only)\n")
	}

	return nil
}

// Certificate returns the CA certificate
func (ca *CA) Certificate() *x509.Certificate {
	return ca.cert
}

// CertificatePEM returns the CA certificate in PEM format
func (ca *CA) CertificatePEM() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.cert.Raw,
	})
}

// PrivateKeyPEM returns the CA private key in PEM format
func (ca *CA) PrivateKeyPEM() []byte {
	ca.mutex.RLock()
	defer ca.mutex.RUnlock()

	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(ca.privateKey),
	})
}

// IssueServiceCertificate generates a certificate for a service based on the request
func (ca *CA) IssueServiceCertificate(req CertRequest) (*CertResponse, error) {
	certPEM, keyPEM, err := ca.GenerateCertificate(req.ServiceName, req.ServiceIP, req.Domains)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %w", err)
	}

	caCertPEM := ca.CertificatePEM()

	return &CertResponse{
		Certificate: certPEM,
		PrivateKey:  keyPEM,
		CACert:      string(caCertPEM),
	}, nil
}

// GenerateCertificate creates a new certificate for a service with the specified domains
func (ca *CA) GenerateCertificate(serviceName, serviceIP string, domains []string) (string, string, error) {
	// Delegate to storage for thread-safe generation and storage
	return ca.storage.GenerateAndStore(ca, serviceName, serviceIP, domains)
}

// GetIssuedCertificates returns all issued certificates
func (ca *CA) GetIssuedCertificates() []*IssuedCert {
	certs, err := ca.storage.GetAll()
	if err != nil {
		// Log error but return empty slice to maintain API compatibility
		return []*IssuedCert{}
	}
	return certs
}

// GetCertificateBySerial returns a certificate by its serial number
func (ca *CA) GetCertificateBySerial(serial string) (*IssuedCert, bool) {
	cert, err := ca.storage.GetBySerial(serial)
	if err != nil {
		return nil, false
	}
	return cert, cert != nil
}

// GetCertificateCount returns the number of issued certificates
func (ca *CA) GetCertificateCount() int {
	count, err := ca.storage.Count()
	if err != nil {
		return 0
	}
	return count
}

// GetCAInfo returns information about the CA certificate
func (ca *CA) GetCAInfo() map[string]interface{} {
	return map[string]interface{}{
		"subject":     ca.cert.Subject.CommonName,
		"valid_until": ca.cert.NotAfter.Format(time.RFC3339),
		"issued_at":   ca.cert.NotBefore.Format(time.RFC3339),
		"serial":      ca.cert.SerialNumber.String(),
	}
}

// ParseCertRequest parses a JSON certificate request
func ParseCertRequest(data []byte) (*CertRequest, error) {
	var req CertRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse certificate request: %w", err)
	}

	// Validate required fields
	if req.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}

	if len(req.Domains) == 0 {
		return nil, fmt.Errorf("domains are required")
	}

	return &req, nil
}

// MarshalCertResponse converts a certificate response to JSON
func MarshalCertResponse(resp *CertResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// saveCAKeyToDisk saves the CA certificate and private key to disk
func (ca *CA) saveCAKeyToDisk() error {
	if ca.persistDir == "" {
		return nil // RAM-only mode
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(ca.persistDir, 0755); err != nil {
		return fmt.Errorf("failed to create persist directory %s: %w", ca.persistDir, err)
	}

	// Save CA certificate
	caCertPath := filepath.Join(ca.persistDir, "ca-cert.pem")
	caCertPEM := ca.CertificatePEM()
	if err := os.WriteFile(caCertPath, caCertPEM, 0644); err != nil {
		return fmt.Errorf("failed to save CA certificate: %w", err)
	}

	// Save CA private key
	caKeyPath := filepath.Join(ca.persistDir, "ca-key.pem")
	caKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(ca.privateKey),
	})
	if err := os.WriteFile(caKeyPath, caKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to save CA private key: %w", err)
	}

	return nil
}

// loadCAFromDisk loads the CA certificate and private key from disk
func (ca *CA) loadCAFromDisk() error {
	if ca.persistDir == "" {
		return nil // RAM-only mode
	}

	caCertPath := filepath.Join(ca.persistDir, "ca-cert.pem")
	caKeyPath := filepath.Join(ca.persistDir, "ca-key.pem")

	// Check if both files exist
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		return nil // No existing CA to load
	}
	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		return nil // No existing CA to load
	}

	// Load certificate
	certPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("failed to decode CA certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Load private key
	keyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read CA private key: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("failed to decode CA private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA private key: %w", err)
	}

	// Set the loaded certificate and key
	ca.mutex.Lock()
	ca.cert = cert
	ca.privateKey = privateKey
	ca.mutex.Unlock()

	return nil
}
