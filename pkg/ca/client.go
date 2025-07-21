// Package ca provides Certificate Authority client functionality
package ca

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GetServiceCertificate requests a certificate from a CA server
// This is a client function that makes HTTP requests to a CA daemon
func GetServiceCertificate(serviceName, serviceIP string, domains []string) (*CertResponse, error) {
	// Default CA daemon address
	caURL := "http://localhost:8090/cert"
	
	// Create certificate request
	req := CertRequest{
		ServiceName: serviceName,
		ServiceIP:   serviceIP,
		Domains:     domains,
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", caURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Make request with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to request certificate from CA: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CA returned status %d", resp.StatusCode)
	}

	// Parse response
	var certResp CertResponse
	if err := json.NewDecoder(resp.Body).Decode(&certResp); err != nil {
		return nil, fmt.Errorf("failed to parse certificate response: %w", err)
	}

	return &certResp, nil
}
