// SPDX-License-Identifier: CC0-1.0

package ca

import (
	"crypto/tls"
	"crypto/x509"
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
)

// UpdateTransport configures the default HTTP client to trust a CA certificate.
// Uses the SGL_CA environment variable to determine the CA server URL.
func UpdateTransport() error {
	caURL := util.MustGetEnv("SGL_CA", "")
	if caURL == "" {
		return ErrNoCAURL
	}

	resp, err := http.Get(caURL + "/ca")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCARequest, err)
	}
	defer resp.Body.Close()

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
