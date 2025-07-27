// SPDX-License-Identifier: CC0-1.0

package util

import (
	"os"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Version should follow semantic versioning pattern (without 'v' prefix)
	if len(Version) < 5 {
		t.Errorf("Version %q should follow X.Y.Z format", Version)
	}
}

func TestMustGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		envValue string
		expected string
	}{
		{
			name:     "environment variable exists",
			key:      "TEST_VAR",
			fallback: "fallback",
			envValue: "actual_value",
			expected: "actual_value",
		},
		{
			name:     "environment variable does not exist",
			key:      "NONEXISTENT_VAR",
			fallback: "fallback_value",
			envValue: "",
			expected: "fallback_value",
		},
		{
			name:     "environment variable is empty",
			key:      "EMPTY_VAR",
			fallback: "fallback_value",
			envValue: "",
			expected: "fallback_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := MustGetEnv(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("MustGetEnv(%q, %q) = %q, want %q", tt.key, tt.fallback, result, tt.expected)
			}
		})
	}
}
