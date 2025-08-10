// SPDX-License-Identifier: CC0-1.0

package ca

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nzions/sharedgolibs/pkg/util"
)

// CreateSecureHTTPSServerV2 creates an HTTPS server with certificates from the CA using the V2 API.
// This is the recommended method for new code that uses simplified SAN-based certificate requests.
//
// Environment Variables Used:
//   - SGL_CA (required): CA server URL for certificate requests
//   - SGL_CA_API_KEY (optional): API key for CA server authentication
//
// Parameters:
//   - serviceName: Name of the service for certificate generation
//   - port: Port number the server will listen on (without ":")
//   - sans: Subject Alternative Names (hostnames and IP addresses)
//   - handler: HTTP handler for the server
//
// Returns a configured *SecureHTTPSServer with TLS certificates, ready to call ListenAndServeTLS().
func CreateSecureHTTPSServerV2(serviceName, port string, sans []string, handler http.Handler) (*SecureHTTPSServer, error) {
	// Request certificate from CA using V2 API
	certResp, err := RequestCertificateV2(serviceName, sans)
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
