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
	"net"
	"strings"
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
	certStore  map[string]*IssuedCert
	mutex      sync.RWMutex
}

// IssuedCert represents a certificate that has been issued by the CA
type IssuedCert struct {
	ServiceName  string    `json:"service_name"`
	Domains      []string  `json:"domains"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Certificate  string    `json:"certificate"`
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
	}
}

// NewCA creates and initializes a new Certificate Authority
func NewCA(config *CAConfig) (*CA, error) {
	if config == nil {
		config = DefaultCAConfig()
	}

	ca := &CA{
		certStore: make(map[string]*IssuedCert),
	}

	if err := ca.initialize(config); err != nil {
		return nil, fmt.Errorf("failed to initialize CA: %w", err)
	}

	return ca, nil
}

// initialize sets up the CA certificate and private key
func (ca *CA) initialize(config *CAConfig) error {
	// Generate CA private key
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
	// Generate service private key
	serviceKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate service private key: %w", err)
	}

	// Generate serial number
	serialNumber := big.NewInt(time.Now().Unix())

	// Create service certificate template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Province:           []string{"Local"},
			Locality:           []string{"Local"},
			Organization:       []string{"SharedGoLibs Development"},
			OrganizationalUnit: []string{"Service"},
			CommonName:         serviceName + ".local",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(30 * 24 * time.Hour), // 30 days
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		SubjectKeyId: []byte{1, 2, 3, 4, 5},
	}

	// Add IP address
	if serviceIP != "" {
		if ip := net.ParseIP(serviceIP); ip != nil {
			template.IPAddresses = []net.IP{ip}
		}
	}

	// Add DNS names
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			// Check if it's an IP address
			if ip := net.ParseIP(domain); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, domain)
			}
		}
	}

	// Create service certificate
	serviceCertDER, err := x509.CreateCertificate(rand.Reader, &template, ca.cert, &serviceKey.PublicKey, ca.privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create service certificate: %w", err)
	}

	// Encode certificate as PEM
	serviceCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serviceCertDER,
	})

	// Encode private key as PEM
	serviceKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serviceKey),
	})

	// Store in certificate store
	issuedCert := &IssuedCert{
		ServiceName:  serviceName,
		Domains:      domains,
		IssuedAt:     time.Now(),
		ExpiresAt:    template.NotAfter,
		Certificate:  string(serviceCertPEM),
		SerialNumber: fmt.Sprintf("%x", serialNumber),
	}

	ca.mutex.Lock()
	ca.certStore[issuedCert.SerialNumber] = issuedCert
	ca.mutex.Unlock()

	return string(serviceCertPEM), string(serviceKeyPEM), nil
}

// GetIssuedCertificates returns all issued certificates
func (ca *CA) GetIssuedCertificates() []*IssuedCert {
	ca.mutex.RLock()
	defer ca.mutex.RUnlock()

	certs := make([]*IssuedCert, 0, len(ca.certStore))
	for _, cert := range ca.certStore {
		certs = append(certs, cert)
	}

	return certs
}

// GetCertificateBySerial returns a certificate by its serial number
func (ca *CA) GetCertificateBySerial(serial string) (*IssuedCert, bool) {
	ca.mutex.RLock()
	defer ca.mutex.RUnlock()

	cert, exists := ca.certStore[serial]
	return cert, exists
}

// GetCertificateCount returns the number of issued certificates
func (ca *CA) GetCertificateCount() int {
	ca.mutex.RLock()
	defer ca.mutex.RUnlock()

	return len(ca.certStore)
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
