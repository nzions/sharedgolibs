// SPDX-License-Identifier: CC0-1.0

package ca

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"

	"github.com/nzions/sharedgolibs/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCARequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: HTTP %d", ErrCARequest, resp.StatusCode)
	}

	// Read the CA certificate
	caCertPEM, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCAResponse, err)
	}

	// Parse the PEM-encoded CA certificate
	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		return fmt.Errorf("%w: failed to parse PEM block", ErrCertParse)
	}

	caCertParsed, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCertParse, err)
	}

	// Create a new certificate pool and add the CA certificate
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCertParsed)

	// Create a custom transport that trusts the CA certificate
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	// Replace the default HTTP client's transport
	http.DefaultClient.Transport = transport

	return nil
}

// RequestCertificate requests a certificate from the CA server for the given service
func RequestCertificate(serviceName, serviceIP string, domains []string) (*CertResponse, error) {
	caURL := util.MustGetEnv("SGL_CA", "")
	if caURL == "" {
		return nil, ErrNoCAURL
	}

	// Create certificate request
	certReq := &CertRequest{
		ServiceName: serviceName,
		ServiceIP:   serviceIP,
		Domains:     domains,
	}

	// Create HTTP request
	req, err := createCertRequest(caURL+"/cert", certReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCARequest, err)
	}

	// Add API key if configured
	apiKey := util.MustGetEnv("SGL_CA_API_KEY", "")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	// Make the request
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

// CreateSecureGRPCServer creates a gRPC server with certificates from the CA.
// This is a convenience method that requests certificates and returns a configured server.
func CreateSecureGRPCServer(serviceName, serviceIP string, domains []string, opts ...grpc.ServerOption) (*grpc.Server, error) {
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

	// Create TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
	})

	// Add TLS credentials to the options
	opts = append(opts, grpc.Creds(creds))

	// Create gRPC server
	server := grpc.NewServer(opts...)

	return server, nil
}

// CreateGRPCCredentials returns gRPC TLS credentials using CA certificates.
// This is a convenience method for clients that need to connect to gRPC servers with CA-issued certificates.
func CreateGRPCCredentials() (credentials.TransportCredentials, error) {
	caURL := util.MustGetEnv("SGL_CA", "")
	if caURL == "" {
		return nil, ErrNoCAURL
	}

	// Create request to get CA certificate
	req, err := http.NewRequest("GET", caURL+"/ca", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCARequest, err)
	}

	// Add API key if configured
	apiKey := util.MustGetEnv("SGL_CA_API_KEY", "")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCARequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrCARequest, resp.StatusCode)
	}

	// Read the CA certificate
	caCertPEM, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCAResponse, err)
	}

	// Parse the PEM-encoded CA certificate
	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		return nil, fmt.Errorf("%w: failed to parse PEM block", ErrCertParse)
	}

	caCertParsed, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCertParse, err)
	}

	// Create a new certificate pool and add the CA certificate
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCertParsed)

	// Create TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		RootCAs: caCertPool,
	})

	return creds, nil
}

// UpdateGRPCDialOptions updates default gRPC dial options to trust CA certificates.
// This is a convenience method for clients that need to dial gRPC servers with CA-issued certificates.
func UpdateGRPCDialOptions() ([]grpc.DialOption, error) {
	// Get gRPC credentials
	creds, err := CreateGRPCCredentials()
	if err != nil {
		return nil, err
	}

	return []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}, nil
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
