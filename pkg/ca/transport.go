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
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/nzions/sharedgolibs/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Transport error types for UpdateTransport function
var (
	// ErrNoCAURL is returned when SGL_CA environment variable is not set
	ErrNoCAURL = fmt.Errorf("SGL_CA environment variable not set")

	// ErrInvalidCAURL is returned when SGL_CA environment variable is not a valid URL
	ErrInvalidCAURL = fmt.Errorf("SGL_CA environment variable is not a valid URL")

	// ErrUnsupportedScheme is returned when SGL_CA URL uses an unsupported scheme
	ErrUnsupportedScheme = fmt.Errorf("SGL_CA URL must use http or https scheme")

	// ErrCARequest is returned when the HTTP request to the CA server fails
	ErrCARequest = fmt.Errorf("failed to request CA certificate")

	// ErrCAResponse is returned when reading the CA response body fails
	ErrCAResponse = fmt.Errorf("failed to read CA certificate response")

	// ErrCertParse is returned when the CA certificate cannot be parsed
	ErrCertParse = fmt.Errorf("failed to parse CA certificate")

	// ErrUnauthorized is returned when API key authentication fails
	ErrUnauthorized = fmt.Errorf("unauthorized: invalid or missing API key")

	// ErrEmulatorMode is returned when Google Cloud emulator environment variables are detected
	ErrEmulatorMode = fmt.Errorf("emulator mode detected: transport update not allowed when Google Cloud emulators are active")
)

// validateCAURL validates that the CA URL is properly formatted
func validateCAURL(caURL string) error {
	if caURL == "" {
		return ErrNoCAURL
	}

	// Parse the URL to validate format
	parsedURL, err := url.Parse(caURL)
	if err != nil {
		return fmt.Errorf("%w: %v (URL: %q)", ErrInvalidCAURL, err, caURL)
	}

	// Check that scheme is http or https
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("%w: got %q, expected http or https (URL: %q)", ErrUnsupportedScheme, parsedURL.Scheme, caURL)
	}

	// Check that host is present
	if parsedURL.Host == "" {
		return fmt.Errorf("%w: missing host (URL: %q)", ErrInvalidCAURL, caURL)
	}

	return nil
}

// checkForEmulatorEnvVars checks for Google Cloud emulator environment variables
// and returns an error if any are found. This prevents transport updates when
// running in emulator mode where CA certificates should not be modified.
func checkForEmulatorEnvVars() error {
	emulatorVars := []string{
		"STORAGE_EMULATOR_HOST",
		"PUBSUB_EMULATOR_HOST",
		"FIRESTORE_EMULATOR_HOST",
		"FIREBASE_EMULATOR_HOST",
		"DATASTORE_EMULATOR_HOST",
		"SPANNER_EMULATOR_HOST",
		"BIGTABLE_EMULATOR_HOST",
		"CLOUD_SQL_EMULATOR_HOST",
		"CLOUDSQL_EMULATOR_HOST",
		"EVENTARC_EMULATOR_HOST",
		"TASKS_EMULATOR_HOST",
		"SECRETMANAGER_EMULATOR_HOST",
		"LOGGING_EMULATOR_HOST",
	}

	for _, envVar := range emulatorVars {
		if value := util.MustGetEnv(envVar, ""); value != "" {
			return fmt.Errorf("%w: %s=%s", ErrEmulatorMode, envVar, value)
		}
	}

	return nil
}

// getValidatedCAURL gets the CA URL from the SGL_CA environment variable and validates it.
// Returns an error if SGL_CA is not set, empty, or contains an invalid URL format.
// The URL must use http:// or https:// scheme and include a valid host.
func getValidatedCAURL() (string, error) {
	caURL := util.MustGetEnv("SGL_CA", "")
	if err := validateCAURL(caURL); err != nil {
		return "", err
	}
	return caURL, nil
}

// UpdateTransport configures the default HTTP client to trust a CA certificate by
// fetching the CA certificate from a CA server and adding it to the trusted root CAs.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL (must be http:// or https://)
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Environment Variables Checked (will error if found):
//   - STORAGE_EMULATOR_HOST, PUBSUB_EMULATOR_HOST, FIRESTORE_EMULATOR_HOST, etc.
//     (Google Cloud emulator environment variables)
//
// Global Variables Modified:
//   - http.DefaultClient.Transport: Replaced with custom transport trusting the CA
//   - http.DefaultTransport: Replaced with the same custom transport
//
// This ensures that both direct usage of http.DefaultClient and libraries that
// create HTTP clients based on http.DefaultTransport will trust the CA certificate.
//
// Returns an error if SGL_CA is not set, invalid, Google Cloud emulator variables
// are detected, or if the CA certificate cannot be fetched or parsed.
func UpdateTransport() error {
	// Check for Google Cloud emulator environment variables
	if err := checkForEmulatorEnvVars(); err != nil {
		return err
	}

	caURL, err := getValidatedCAURL()
	if err != nil {
		return err
	}

	return updateTransportWithCA(caURL)
}

// UpdateTransportOnlyIf configures the default HTTP client to trust a CA certificate
// only if the SGL_CA environment variable is set. This is a conditional version of
// UpdateTransport that gracefully handles the case where no CA server is configured.
//
// Environment Variables Used:
//   - SGL_CA (optional): CA server URL (must be http:// or https://) - if not set, function returns nil
//   - SGL_CA_API_KEY (optional): API key for CA server authentication (only used if SGL_CA is set)
//
// Environment Variables Checked (will error if found and SGL_CA is set):
//   - STORAGE_EMULATOR_HOST, PUBSUB_EMULATOR_HOST, FIRESTORE_EMULATOR_HOST, etc.
//     (Google Cloud emulator environment variables)
//
// Global Variables Modified (only if SGL_CA is set):
//   - http.DefaultClient.Transport: Replaced with custom transport trusting the CA
//   - http.DefaultTransport: Replaced with the same custom transport
//
// Returns nil without error if SGL_CA is not set (no-op).
// Returns an error if SGL_CA is set but invalid, Google Cloud emulator variables
// are detected, or if the CA certificate cannot be fetched or parsed.
func UpdateTransportOnlyIf() error {
	caURL := util.MustGetEnv("SGL_CA", "")
	if caURL == "" {
		// SGL_CA is not set, do nothing
		return nil
	}

	// Check for Google Cloud emulator environment variables
	if err := checkForEmulatorEnvVars(); err != nil {
		return err
	}

	// Validate the URL since it's set
	if err := validateCAURL(caURL); err != nil {
		return err
	}

	slog.Info("Updating HTTP transport to trust CA", "url", caURL)
	return updateTransportWithCA(caURL)
}

// UpdateTransportMust configures the default HTTP client to trust a CA certificate
// and panics if the operation fails. This is a convenience function for scenarios
// where transport update failure should be treated as a fatal error.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL (must be http:// or https://)
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Environment Variables Checked (will panic if found):
//   - STORAGE_EMULATOR_HOST, PUBSUB_EMULATOR_HOST, FIRESTORE_EMULATOR_HOST, etc.
//     (Google Cloud emulator environment variables)
//
// Global Variables Modified:
//   - http.DefaultClient.Transport: Replaced with custom transport trusting the CA
//   - http.DefaultTransport: Replaced with the same custom transport
//
// Panics if SGL_CA is not set, invalid, Google Cloud emulator variables
// are detected, or if the CA certificate cannot be fetched or parsed.
func UpdateTransportMust() {
	if err := UpdateTransport(); err != nil {
		panic(fmt.Sprintf("failed to update transport: %v", err))
	}
}

// updateTransportWithCA handles the actual transport update logic by fetching
// the CA certificate from the specified URL and configuring global HTTP transports.
//
// This function:
//  1. Makes a GET request to caURL+"/ca" to fetch the CA certificate
//  2. Optionally includes SGL_CA_API_KEY header if the environment variable is set
//  3. Parses the returned PEM-encoded CA certificate
//  4. Creates a new http.Transport with a TLS config trusting the CA
//  5. Replaces both http.DefaultClient.Transport and http.DefaultTransport
//
// Global Variables Modified:
//   - http.DefaultClient.Transport: Set to new transport with CA trust
//   - http.DefaultTransport: Set to the same transport instance
//
// Parameters:
//   - caURL: The base URL of the CA server (without "/ca" path)
//
// Returns an error if the HTTP request fails, the response is invalid,
// or the certificate cannot be parsed.
func updateTransportWithCA(caURL string) error {
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

	// Replace BOTH the default HTTP client's transport AND the default transport
	// This ensures that both direct usage of http.DefaultClient and libraries that
	// create clients based on http.DefaultTransport will trust our CA
	http.DefaultClient.Transport = transport
	http.DefaultTransport = transport

	return nil
}

// RequestCertificateV2 requests a certificate from the CA server using a simplified API.
// This function automatically detects IP addresses in the SANs list and properly
// configures the certificate with the first non-IP SAN as the Common Name.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL (must be http:// or https://)
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Parameters:
//   - serviceName: Name of the service requesting the certificate
//   - sans: Subject Alternative Names (hostnames and IP addresses)
//
// Returns a CertResponse containing the PEM-encoded certificate and private key,
// or an error if the request fails or authentication is required but invalid.
func RequestCertificateV2(serviceName string, sans []string) (*CertResponse, error) {
	caURL, err := getValidatedCAURL()
	if err != nil {
		return nil, err
	}

	// Create certificate request with new V2 format
	certReq := &CertRequestV2{
		ServiceName: serviceName,
		SANs:        sans,
	}

	// Create HTTP request
	req, err := createCertRequestV2(caURL+"/cert", certReq)
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

// RequestCertificate requests a certificate from the CA server for the given service.
// The certificate includes the service name, IP address, and additional domain names.
//
// DEPRECATED: Use RequestCertificateV2 for the simplified API with automatic IP detection.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL (must be http:// or https://)
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Parameters:
//   - serviceName: Name of the service requesting the certificate
//   - serviceIP: IP address of the service
//   - domains: Additional domain names to include in the certificate
//
// Returns a CertResponse containing the PEM-encoded certificate and private key,
// or an error if the request fails or authentication is required but invalid.
func RequestCertificate(serviceName, serviceIP string, domains []string) (*CertResponse, error) {
	caURL, err := getValidatedCAURL()
	if err != nil {
		return nil, err
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
// This is a convenience method that requests certificates from the CA server and
// returns a configured SecureHTTPSServer ready to serve HTTPS traffic.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL for certificate requests
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Parameters:
//   - serviceName: Name of the service for certificate generation
//   - serviceIP: IP address of the service
//   - port: Port number the server will listen on (without ":")
//   - domains: Additional domain names to include in the certificate
//   - handler: HTTP handler for the server
//
// Returns a configured *SecureHTTPSServer with TLS certificates, ready to call ListenAndServeTLS().
func CreateSecureHTTPSServer(serviceName, serviceIP, port string, domains []string, handler http.Handler) (*SecureHTTPSServer, error) {
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

	return NewSecureHTTPSServer(server), nil
}

// CreateSecureGRPCServer creates a gRPC server with certificates from the CA.
// This is a convenience method that requests certificates from the CA server and
// returns a configured gRPC server with TLS transport credentials.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL for certificate requests
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Parameters:
//   - serviceName: Name of the service for certificate generation
//   - serviceIP: IP address of the service
//   - domains: Additional domain names to include in the certificate
//   - opts: Additional gRPC server options (TLS credentials will be appended)
//
// Returns a configured *grpc.Server with TLS credentials, ready to serve.
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
// This is a convenience method for gRPC clients that need to connect to servers
// with CA-issued certificates. The credentials include the CA certificate in the
// trusted root CAs, allowing verification of server certificates issued by the CA.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL to fetch the CA certificate from
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Returns credentials.TransportCredentials that can be used with grpc.WithTransportCredentials()
// for secure gRPC client connections.
func CreateGRPCCredentials() (credentials.TransportCredentials, error) {
	caURL, err := getValidatedCAURL()
	if err != nil {
		return nil, err
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

// UpdateGRPCDialOptions returns configured gRPC dial options to trust CA certificates.
// This is a convenience method for gRPC clients that need to dial servers with CA-issued
// certificates. The returned dial options include TLS transport credentials with the
// CA certificate in the trusted root CAs.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL to fetch the CA certificate from
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Returns a slice of grpc.DialOption that can be passed to grpc.Dial() or grpc.NewClient()
// to establish secure connections to gRPC servers with CA-issued certificates.
//
// Example:
//
//	opts, err := ca.UpdateGRPCDialOptions()
//	if err != nil { return err }
//	conn, err := grpc.Dial("server:443", opts...)
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

// createCertRequest creates an HTTP POST request for certificate generation.
// Serializes the CertRequest struct to JSON and sets appropriate headers.
//
// Parameters:
//   - url: The full URL endpoint for certificate requests (typically caURL+"/cert")
//   - certReq: The certificate request data to be JSON-encoded
//
// Returns an *http.Request ready to be executed, with Content-Type set to application/json.
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

// createCertRequestV2 creates an HTTP POST request for certificate generation using V2 format.
// Serializes any certificate request struct to JSON and sets appropriate headers.
//
// Parameters:
//   - url: The full URL endpoint for certificate requests (typically caURL+"/cert")
//   - certReq: The certificate request data to be JSON-encoded (CertRequest or CertRequestV2)
//
// Returns an *http.Request ready to be executed, with Content-Type set to application/json.
func createCertRequestV2(url string, certReq interface{}) (*http.Request, error) {
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
