// SPDX-License-Identifier: CC0-1.0

package ca

import (
	"net/http"
	"testing"

	"github.com/nzions/sharedgolibs/pkg/logi"
)

func TestCreateSecureDualProtocolServer(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		port        string
		sans        []string
		handler     http.Handler
		logger      logi.Logger
		wantErr     bool
		errContains string
	}{
		{
			name:        "invalid service name",
			serviceName: "",
			port:        "8443",
			sans:        []string{"localhost"},
			handler:     nil,
			logger:      nil,
			wantErr:     true,
			errContains: "failed to request certificate",
		},
		{
			name:        "valid parameters without CA",
			serviceName: "test-service",
			port:        "8443",
			sans:        []string{"example.com", "127.0.0.1", "192.168.1.100"},
			handler:     nil,
			logger:      nil,
			wantErr:     true,
			errContains: "failed to request certificate",
		},
		{
			name:        "mixed IPs and hostnames",
			serviceName: "mixed-service",
			port:        "8443",
			sans:        []string{"api.example.com", "127.0.0.1", "192.168.1.100", "localhost"},
			handler:     nil,
			logger:      nil,
			wantErr:     true,
			errContains: "failed to request certificate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := CreateSecureDualProtocolServer(tt.serviceName, tt.port, tt.sans, tt.handler, tt.logger)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateSecureDualProtocolServer() expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("CreateSecureDualProtocolServer() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CreateSecureDualProtocolServer() unexpected error = %v", err)
				return
			}

			if server == nil {
				t.Error("CreateSecureDualProtocolServer() returned nil server")
			}
		})
	}
} // Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(substr) == 0 || indexOfString(s, substr) >= 0)
}

// Helper function to find the index of a substring
func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
