// SPDX-License-Identifier: CC0-1.0

package ca

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nzions/sharedgolibs/pkg/util"
)

// Transport error types for UpdateTransport function
var (
	// ErrNoCAURL is returned when SGL_CA environment variable is not set
	ErrNoCAURL = fmt.Errorf("SGL_CA environment variable not set")

	// ErrCARequest is returned when the HTTP request to the CA server fails
	ErrCARequest = fmt.Errorf("failed to request CA certificate")

	// ErrCAResponse is returned when reading the CA response body fails
	ErrCAResponse = fmt.Errorf("failed to read CA certificate response")

	// ErrCertParse is returned when the CA certificate cannot be parsed
	ErrCertParse = fmt.Errorf("failed to parse CA certificate")

	// ErrUnauthorized is returned when API key authentication fails
	ErrUnauthorized = fmt.Errorf("unauthorized: invalid or missing API key")
)

// UpdateTransport configures the default HTTP client to trust a CA certificate.
// Uses the SGL_CA environment variable to determine the CA server URL.
// Optionally uses SGL_CA_API_KEY environment variable for authentication.
func UpdateTransport() error {
	caURL := util.MustGetEnv("SGL_CA", "")
	if caURL == "" {
		return ErrNoCAURL
	}

	// Create request with optional API key
	req, err := http.NewRequest("GET", caURL+"/ca", nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCARequest, err)
	}

	// Add API key if configured
	apiKey := util.MustGetEnv("SGL_CA_API_KEY", "")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCARequest, err)
	}
	defer resp.Body.Close()

	// Check for unauthorized response
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: server returned status %d", ErrCARequest, resp.StatusCode)
	}

	caCertPEM, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCAResponse, err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertPEM) {
		return ErrCertParse
	}

	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: certPool},
	}

	return nil
}

// RequestCertificate requests a certificate from the CA server.
// Uses the SGL_CA and optionally SGL_CA_API_KEY environment variables.
func RequestCertificate(serviceName, serviceIP string, domains []string) (*CertResponse, error) {
	caURL := util.MustGetEnv("SGL_CA", "")
	if caURL == "" {
		return nil, ErrNoCAURL
	}

	// Create the certificate request
	certReq := CertRequest{
		ServiceName: serviceName,
		ServiceIP:   serviceIP,
		Domains:     domains,
	}

	// Create HTTP request
	req, err := createCertRequest(caURL+"/cert", &certReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCARequest, err)
	}

	// Add API key if configured
	apiKey := util.MustGetEnv("SGL_CA_API_KEY", "")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCARequest, err)
	}
	defer resp.Body.Close()

	// Check for unauthorized response
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: server returned status %d", ErrCARequest, resp.StatusCode)
	}

	var certResp CertResponse
	if err := json.NewDecoder(resp.Body).Decode(&certResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCAResponse, err)
	}

	return &certResp, nil
}

// CreateSecureHTTPSServer creates an HTTPS server with certificates from the CA.
// This is a convenience method that requests certificates and returns a configured server.
func CreateSecureHTTPSServer(serviceName, serviceIP, port string, domains []string, handler http.Handler) (*http.Server, error) {
	// Request certificate from CA
	certResp, err := RequestCertificate(serviceName, serviceIP, domains)
	if err != nil {
		return nil, fmt.Errorf("failed to request certificate: %w", err)
	}

	// Parse the certificate and key
	cert, err := tls.X509KeyPair([]byte(certResp.Certificate), []byte(certResp.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Create HTTPS server
	server := &http.Server{
		Addr:      ":" + port,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	return server, nil
}

// createCertRequest creates a JSON request for certificate generation
func createCertRequest(url string, certReq *CertRequest) (*http.Request, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(certReq); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
