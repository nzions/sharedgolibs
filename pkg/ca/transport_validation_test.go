package ca

import (
	"errors"
	"testing"
)

func TestValidateCAURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
		errorType error
	}{
		{
			name:      "Valid HTTP URL",
			url:       "http://ca:8090",
			wantError: false,
		},
		{
			name:      "Valid HTTPS URL",
			url:       "https://ca.example.com:8090",
			wantError: false,
		},
		{
			name:      "Valid HTTP URL with localhost",
			url:       "http://localhost:8090",
			wantError: false,
		},
		{
			name:      "Valid HTTPS URL with IP",
			url:       "https://192.168.1.100:8090",
			wantError: false,
		},
		{
			name:      "Empty URL",
			url:       "",
			wantError: true,
			errorType: ErrNoCAURL,
		},
		{
			name:      "Invalid URL format",
			url:       "not-a-url",
			wantError: true,
			errorType: ErrUnsupportedScheme, // no scheme means unsupported scheme
		},
		{
			name:      "URL without scheme",
			url:       "ca:8090",
			wantError: true,
			errorType: ErrUnsupportedScheme,
		},
		{
			name:      "URL with unsupported scheme",
			url:       "ftp://ca:8090",
			wantError: true,
			errorType: ErrUnsupportedScheme,
		},
		{
			name:      "URL without host",
			url:       "http://",
			wantError: true,
			errorType: ErrInvalidCAURL,
		},
		{
			name:      "URL with path (should still be valid)",
			url:       "http://ca:8090/some/path",
			wantError: false,
		},
		{
			name:      "URL with query params (should still be valid)",
			url:       "http://ca:8090?param=value",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCAURL(tt.url)

			if tt.wantError {
				if err == nil {
					t.Errorf("validateCAURL() expected error but got none")
					return
				}

				// Check error type if specified
				if tt.errorType != nil {
					if !isExpectedError(err, tt.errorType) {
						t.Errorf("validateCAURL() expected error type %v, got %v", tt.errorType, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("validateCAURL() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetValidatedCAURL(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantError bool
		wantURL   string
	}{
		{
			name:      "Valid HTTP URL",
			envValue:  "http://ca:8090",
			wantError: false,
			wantURL:   "http://ca:8090",
		},
		{
			name:      "Valid HTTPS URL",
			envValue:  "https://ca.example.com:8090",
			wantError: false,
			wantURL:   "https://ca.example.com:8090",
		},
		{
			name:      "Invalid URL format",
			envValue:  "not-a-url",
			wantError: true,
		},
		{
			name:      "Empty URL",
			envValue:  "",
			wantError: true,
		},
		{
			name:      "URL without scheme",
			envValue:  "ca:8090",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			t.Setenv("SGL_CA", tt.envValue)

			url, err := getValidatedCAURL()

			if tt.wantError {
				if err == nil {
					t.Errorf("getValidatedCAURL() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("getValidatedCAURL() unexpected error: %v", err)
					return
				}

				if url != tt.wantURL {
					t.Errorf("getValidatedCAURL() expected URL %q, got %q", tt.wantURL, url)
				}
			}
		})
	}
}

// Helper function to check if error is of expected type
func isExpectedError(err, expectedType error) bool {
	return errors.Is(err, expectedType)
}
