package ca

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNewCA(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	if ca == nil {
		t.Fatal("CA is nil")
	}

	if ca.cert == nil {
		t.Fatal("CA certificate is nil")
	}

	if ca.privateKey == nil {
		t.Fatal("CA private key is nil")
	}

	if ca.storage == nil {
		t.Fatal("Certificate storage is nil")
	}
}

func TestCAWithCustomConfig(t *testing.T) {
	config := &CAConfig{
		Country:            []string{"CA"},
		Province:           []string{"Ontario"},
		Locality:           []string{"Toronto"},
		Organization:       []string{"Test Org"},
		OrganizationalUnit: []string{"Test Unit"},
		CommonName:         "Test CA",
		ValidityPeriod:     30 * 24 * time.Hour, // 30 days
		KeySize:            2048,
	}

	ca, err := NewCA(config)
	if err != nil {
		t.Fatalf("Failed to create CA with custom config: %v", err)
	}

	// Verify the certificate has the expected values
	cert := ca.Certificate()
	if cert.Subject.CommonName != "Test CA" {
		t.Errorf("Expected CommonName 'Test CA', got '%s'", cert.Subject.CommonName)
	}

	if len(cert.Subject.Country) != 1 || cert.Subject.Country[0] != "CA" {
		t.Errorf("Expected Country ['CA'], got %v", cert.Subject.Country)
	}

	if len(cert.Subject.Organization) != 1 || cert.Subject.Organization[0] != "Test Org" {
		t.Errorf("Expected Organization ['Test Org'], got %v", cert.Subject.Organization)
	}

	// Check validity period (approximately 30 days)
	validity := cert.NotAfter.Sub(cert.NotBefore)
	expectedValidity := 30 * 24 * time.Hour
	if validity < expectedValidity-time.Hour || validity > expectedValidity+time.Hour {
		t.Errorf("Expected validity period around %v, got %v", expectedValidity, validity)
	}
}

func TestCertificatePEM(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	pemBytes := ca.CertificatePEM()
	if len(pemBytes) == 0 {
		t.Fatal("PEM bytes are empty")
	}

	// Verify PEM format
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		t.Fatal("Failed to decode PEM block")
	}

	if block.Type != "CERTIFICATE" {
		t.Errorf("Expected PEM type 'CERTIFICATE', got '%s'", block.Type)
	}

	// Verify the decoded certificate matches the original
	decodedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse decoded certificate: %v", err)
	}

	originalCert := ca.Certificate()
	if !decodedCert.Equal(originalCert) {
		t.Error("Decoded certificate does not match original")
	}
}

func TestGenerateCertificate(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	serviceName := "test-service"
	serviceIP := "192.168.1.100"
	domains := []string{"test.example.com", "test2.example.com", "127.0.0.1"}

	certPEM, keyPEM, err := ca.GenerateCertificate(serviceName, serviceIP, domains)
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	if certPEM == "" {
		t.Fatal("Certificate PEM is empty")
	}

	if keyPEM == "" {
		t.Fatal("Private key PEM is empty")
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

	// Verify certificate properties
	// With the new CN selection logic, the first non-IP domain becomes the CN
	expectedCN := "test.example.com" // first non-IP domain from the domains list
	if cert.Subject.CommonName != expectedCN {
		t.Errorf("Expected CommonName '%s', got '%s'", expectedCN, cert.Subject.CommonName)
	}

	// Verify DNS names
	expectedDNS := []string{"test.example.com", "test2.example.com"}
	if len(cert.DNSNames) != len(expectedDNS) {
		t.Errorf("Expected %d DNS names, got %d", len(expectedDNS), len(cert.DNSNames))
	}
	for i, dns := range expectedDNS {
		if i >= len(cert.DNSNames) || cert.DNSNames[i] != dns {
			t.Errorf("Expected DNS name '%s', got '%s'", dns, cert.DNSNames[i])
		}
	}

	// Verify IP addresses
	expectedIPs := []net.IP{net.ParseIP("192.168.1.100"), net.ParseIP("127.0.0.1")}
	if len(cert.IPAddresses) != len(expectedIPs) {
		t.Errorf("Expected %d IP addresses, got %d", len(expectedIPs), len(cert.IPAddresses))
	}
	for i, ip := range expectedIPs {
		if i >= len(cert.IPAddresses) || !cert.IPAddresses[i].Equal(ip) {
			t.Errorf("Expected IP address '%s', got '%s'", ip, cert.IPAddresses[i])
		}
	}

	// Verify the certificate is signed by the CA
	caCert := ca.Certificate()
	err = cert.CheckSignatureFrom(caCert)
	if err != nil {
		t.Errorf("Certificate is not properly signed by CA: %v", err)
	}
}

func TestIssueServiceCertificate(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	req := CertRequest{
		ServiceName: "web-service",
		ServiceIP:   "10.0.0.1",
		Domains:     []string{"web.example.com", "api.example.com"},
	}

	response, err := ca.IssueServiceCertificate(req)
	if err != nil {
		t.Fatalf("Failed to issue service certificate: %v", err)
	}

	if response.Certificate == "" {
		t.Fatal("Certificate in response is empty")
	}

	if response.PrivateKey == "" {
		t.Fatal("Private key in response is empty")
	}

	if response.CACert == "" {
		t.Fatal("CA certificate in response is empty")
	}

	// Verify the CA certificate in response matches our CA
	block, _ := pem.Decode([]byte(response.CACert))
	if block == nil {
		t.Fatal("Failed to decode CA certificate from response")
	}

	responseCACert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse CA certificate from response: %v", err)
	}

	if !responseCACert.Equal(ca.Certificate()) {
		t.Error("CA certificate in response does not match CA certificate")
	}
}

func TestIssueServiceCertificateV2(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Test V2 API with automatic IP detection and CN selection
	req := CertRequestV2{
		ServiceName: "test-service-v2",
		SANs:        []string{"api.example.com", "backup.example.com", "192.168.1.100", "127.0.0.1"},
	}

	response, err := ca.IssueServiceCertificateV2(req)
	if err != nil {
		t.Fatalf("Failed to issue service certificate with V2 API: %v", err)
	}

	if response.Certificate == "" {
		t.Fatal("Certificate in response is empty")
	}

	if response.PrivateKey == "" {
		t.Fatal("Private key in response is empty")
	}

	if response.CACert == "" {
		t.Fatal("CA certificate in response is empty")
	}

	// Parse and verify the certificate
	block, _ := pem.Decode([]byte(response.Certificate))
	if block == nil {
		t.Fatal("Failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Verify CN is the first non-IP SAN (api.example.com)
	expectedCN := "api.example.com"
	if cert.Subject.CommonName != expectedCN {
		t.Errorf("Expected CommonName '%s', got '%s'", expectedCN, cert.Subject.CommonName)
	}

	// Verify DNS names include all non-IP SANs
	expectedDNS := []string{"api.example.com", "backup.example.com"}
	if len(cert.DNSNames) != len(expectedDNS) {
		t.Errorf("Expected %d DNS names, got %d", len(expectedDNS), len(cert.DNSNames))
	}
	for i, expected := range expectedDNS {
		if i >= len(cert.DNSNames) || cert.DNSNames[i] != expected {
			t.Errorf("Expected DNS name %d to be '%s', got '%s'", i, expected, cert.DNSNames[i])
		}
	}

	// Verify IP addresses include all IP SANs
	expectedIPs := []string{"192.168.1.100", "127.0.0.1"}
	if len(cert.IPAddresses) != len(expectedIPs) {
		t.Errorf("Expected %d IP addresses, got %d", len(expectedIPs), len(cert.IPAddresses))
	}
	for i, expectedIP := range expectedIPs {
		if i >= len(cert.IPAddresses) || cert.IPAddresses[i].String() != expectedIP {
			t.Errorf("Expected IP address %d to be '%s', got '%s'", i, expectedIP, cert.IPAddresses[i].String())
		}
	}
}

func TestIssueServiceCertificateV2_IPOnly(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Test V2 API with only IP addresses
	req := CertRequestV2{
		ServiceName: "ip-only-service",
		SANs:        []string{"192.168.1.100", "10.0.0.50"},
	}

	response, err := ca.IssueServiceCertificateV2(req)
	if err != nil {
		t.Fatalf("Failed to issue service certificate with IP-only SANs: %v", err)
	}

	// Parse and verify the certificate
	block, _ := pem.Decode([]byte(response.Certificate))
	if block == nil {
		t.Fatal("Failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Verify CN is the first IP address
	expectedCN := "192.168.1.100"
	if cert.Subject.CommonName != expectedCN {
		t.Errorf("Expected CommonName '%s', got '%s'", expectedCN, cert.Subject.CommonName)
	}

	// Verify no DNS names (only IPs)
	if len(cert.DNSNames) != 0 {
		t.Errorf("Expected 0 DNS names for IP-only certificate, got %d", len(cert.DNSNames))
	}

	// Verify IP addresses
	expectedIPs := []string{"192.168.1.100", "10.0.0.50"}
	if len(cert.IPAddresses) != len(expectedIPs) {
		t.Errorf("Expected %d IP addresses, got %d", len(expectedIPs), len(cert.IPAddresses))
	}
	for i, expectedIP := range expectedIPs {
		if i >= len(cert.IPAddresses) || cert.IPAddresses[i].String() != expectedIP {
			t.Errorf("Expected IP address %d to be '%s', got '%s'", i, expectedIP, cert.IPAddresses[i].String())
		}
	}
}

func TestIssueServiceCertificateV2_EmptySANs(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Test V2 API with empty SANs should return error
	req := CertRequestV2{
		ServiceName: "invalid-service",
		SANs:        []string{},
	}

	_, err = ca.IssueServiceCertificateV2(req)
	if err == nil {
		t.Fatal("Expected error for empty SANs, but got success")
	}

	expectedError := "domains/SANs cannot be empty"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetIssuedCertificates(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Initially should be empty
	certs := ca.GetIssuedCertificates()
	if len(certs) != 0 {
		t.Errorf("Expected 0 issued certificates, got %d", len(certs))
	}

	// Issue a certificate
	_, _, err = ca.GenerateCertificate("test-service", "127.0.0.1", []string{"test.com"})
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// Should now have 1 certificate
	certs = ca.GetIssuedCertificates()
	if len(certs) != 1 {
		t.Errorf("Expected 1 issued certificate, got %d", len(certs))
	}

	cert := certs[0]
	if cert.ServiceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", cert.ServiceName)
	}

	expectedDomains := []string{"test.com"}
	if len(cert.Domains) != len(expectedDomains) {
		t.Errorf("Expected %d domains, got %d", len(expectedDomains), len(cert.Domains))
	}
	if cert.Domains[0] != expectedDomains[0] {
		t.Errorf("Expected domain '%s', got '%s'", expectedDomains[0], cert.Domains[0])
	}
}

func TestGetCertificateBySerial(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Generate a certificate
	_, _, err = ca.GenerateCertificate("test-service", "127.0.0.1", []string{"test.com"})
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	certs := ca.GetIssuedCertificates()
	if len(certs) != 1 {
		t.Fatalf("Expected 1 certificate, got %d", len(certs))
	}

	serial := certs[0].SerialNumber

	// Test getting existing certificate
	cert, exists := ca.GetCertificateBySerial(serial)
	if !exists {
		t.Fatal("Certificate should exist")
	}

	if cert.SerialNumber != serial {
		t.Errorf("Expected serial '%s', got '%s'", serial, cert.SerialNumber)
	}

	// Test getting non-existent certificate
	_, exists = ca.GetCertificateBySerial("nonexistent")
	if exists {
		t.Fatal("Certificate should not exist")
	}
}

func TestGetCertificateCount(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Initially should be 0
	count := ca.GetCertificateCount()
	if count != 0 {
		t.Errorf("Expected 0 certificates, got %d", count)
	}

	// Issue certificates
	for i := 0; i < 3; i++ {
		req := CertRequest{
			ServiceName: "test-service",
			ServiceIP:   "127.0.0.1",
			Domains:     []string{"test.com"},
		}
		_, err = ca.IssueServiceCertificate(req)
		if err != nil {
			t.Fatalf("Failed to issue certificate %d: %v", i, err)
		}
	}

	count = ca.GetCertificateCount()
	if count != 3 {
		t.Errorf("Expected 3 certificates, got %d", count)
	}
}

func TestGetCAInfo(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	info := ca.GetCAInfo()

	if info["subject"] != "SharedGoLibs Root CA" {
		t.Errorf("Expected subject 'SharedGoLibs Root CA', got '%v'", info["subject"])
	}

	if info["valid_until"] == nil {
		t.Error("valid_until should not be nil")
	}

	if info["issued_at"] == nil {
		t.Error("issued_at should not be nil")
	}

	if info["serial"] == nil {
		t.Error("serial should not be nil")
	}
}

func TestParseCertRequest(t *testing.T) {
	// Test valid request
	validJSON := `{
		"service_name": "test-service",
		"service_ip": "192.168.1.1",
		"domains": ["test.com", "api.test.com"]
	}`

	req, err := ParseCertRequest([]byte(validJSON))
	if err != nil {
		t.Fatalf("Failed to parse valid request: %v", err)
	}

	if req.ServiceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", req.ServiceName)
	}

	if req.ServiceIP != "192.168.1.1" {
		t.Errorf("Expected service IP '192.168.1.1', got '%s'", req.ServiceIP)
	}

	expectedDomains := []string{"test.com", "api.test.com"}
	if len(req.Domains) != len(expectedDomains) {
		t.Errorf("Expected %d domains, got %d", len(expectedDomains), len(req.Domains))
	}

	// Test invalid JSON
	_, err = ParseCertRequest([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test missing service name
	invalidJSON := `{
		"service_ip": "192.168.1.1",
		"domains": ["test.com"]
	}`

	_, err = ParseCertRequest([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected error for missing service name")
	}
	if !strings.Contains(err.Error(), "service_name is required") {
		t.Errorf("Expected 'service_name is required' error, got '%v'", err)
	}

	// Test missing domains
	invalidJSON2 := `{
		"service_name": "test-service",
		"service_ip": "192.168.1.1",
		"domains": []
	}`

	_, err = ParseCertRequest([]byte(invalidJSON2))
	if err == nil {
		t.Error("Expected error for missing domains")
	}
	if !strings.Contains(err.Error(), "domains are required") {
		t.Errorf("Expected 'domains are required' error, got '%v'", err)
	}
}

func TestMarshalCertResponse(t *testing.T) {
	response := &CertResponse{
		Certificate: "cert-pem-data",
		PrivateKey:  "key-pem-data",
		CACert:      "ca-cert-pem-data",
	}

	data, err := MarshalCertResponse(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal and verify
	var unmarshaled CertResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshaled.Certificate != response.Certificate {
		t.Errorf("Certificate mismatch: expected '%s', got '%s'", response.Certificate, unmarshaled.Certificate)
	}

	if unmarshaled.PrivateKey != response.PrivateKey {
		t.Errorf("PrivateKey mismatch: expected '%s', got '%s'", response.PrivateKey, unmarshaled.PrivateKey)
	}

	if unmarshaled.CACert != response.CACert {
		t.Errorf("CACert mismatch: expected '%s', got '%s'", response.CACert, unmarshaled.CACert)
	}
}

func TestDefaultCAConfig(t *testing.T) {
	config := DefaultCAConfig()

	if config == nil {
		t.Fatal("Default config is nil")
	}

	if len(config.Country) != 1 || config.Country[0] != "US" {
		t.Errorf("Expected Country ['US'], got %v", config.Country)
	}

	if config.CommonName != "SharedGoLibs Root CA" {
		t.Errorf("Expected CommonName 'SharedGoLibs Root CA', got '%s'", config.CommonName)
	}

	if config.ValidityPeriod != 365*24*time.Hour {
		t.Errorf("Expected ValidityPeriod 1 year, got %v", config.ValidityPeriod)
	}

	if config.KeySize != 4096 {
		t.Errorf("Expected KeySize 4096, got %d", config.KeySize)
	}
}

func TestConcurrentCertificateGeneration(t *testing.T) {
	ca, err := NewCA(nil)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	const numGoroutines = 10
	const certsPerGoroutine = 5

	done := make(chan error, numGoroutines)

	// Start multiple goroutines generating certificates
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < certsPerGoroutine; j++ {
				serviceName := fmt.Sprintf("service-%d-%d", goroutineID, j)
				domains := []string{fmt.Sprintf("%s.example.com", serviceName)}

				req := CertRequest{
					ServiceName: serviceName,
					ServiceIP:   "127.0.0.1",
					Domains:     domains,
				}
				_, err := ca.IssueServiceCertificate(req)
				if err != nil {
					done <- fmt.Errorf("goroutine %d cert %d failed: %v", goroutineID, j, err)
					return
				}
			}
			done <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}

	// Verify all certificates were generated
	expectedCount := numGoroutines * certsPerGoroutine
	actualCount := ca.GetCertificateCount()
	if actualCount != expectedCount {
		t.Errorf("Expected %d certificates, got %d", expectedCount, actualCount)
	}

	// Verify all certificates are unique
	certs := ca.GetIssuedCertificates()
	serials := make(map[string]bool)
	for _, cert := range certs {
		if serials[cert.SerialNumber] {
			t.Errorf("Duplicate serial number found: %s", cert.SerialNumber)
		}
		serials[cert.SerialNumber] = true
	}
}
