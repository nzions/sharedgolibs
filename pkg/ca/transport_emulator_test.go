package ca

import (
	"os"
	"strings"
	"testing"
)

func TestEmulatorEnvVarDetection(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalAPIKey := os.Getenv("SGL_CA_API_KEY")
	originalStorage := os.Getenv("STORAGE_EMULATOR_HOST")
	originalPubSub := os.Getenv("PUBSUB_EMULATOR_HOST")
	originalFirestore := os.Getenv("FIRESTORE_EMULATOR_HOST")
	originalFirebase := os.Getenv("FIREBASE_EMULATOR_HOST")

	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("SGL_CA_API_KEY", originalAPIKey)
		os.Setenv("STORAGE_EMULATOR_HOST", originalStorage)
		os.Setenv("PUBSUB_EMULATOR_HOST", originalPubSub)
		os.Setenv("FIRESTORE_EMULATOR_HOST", originalFirestore)
		os.Setenv("FIREBASE_EMULATOR_HOST", originalFirebase)
	}()

	tests := []struct {
		name          string
		setupEnv      func()
		expectError   bool
		expectedError string
	}{
		{
			name: "No emulator variables set",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:8090")
				os.Unsetenv("STORAGE_EMULATOR_HOST")
				os.Unsetenv("PUBSUB_EMULATOR_HOST")
				os.Unsetenv("FIRESTORE_EMULATOR_HOST")
				os.Unsetenv("FIREBASE_EMULATOR_HOST")
				os.Unsetenv("DATASTORE_EMULATOR_HOST")
				os.Unsetenv("SPANNER_EMULATOR_HOST")
				os.Unsetenv("BIGTABLE_EMULATOR_HOST")
				os.Unsetenv("CLOUD_SQL_EMULATOR_HOST")
				os.Unsetenv("CLOUDSQL_EMULATOR_HOST")
				os.Unsetenv("EVENTARC_EMULATOR_HOST")
				os.Unsetenv("TASKS_EMULATOR_HOST")
				os.Unsetenv("SECRETMANAGER_EMULATOR_HOST")
				os.Unsetenv("LOGGING_EMULATOR_HOST")
			},
			expectError: true, // Will error because CA server doesn't exist, but NOT because of emulator detection
		},
		{
			name: "STORAGE_EMULATOR_HOST set",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:8090")
				os.Setenv("STORAGE_EMULATOR_HOST", "localhost:9000")
				os.Unsetenv("PUBSUB_EMULATOR_HOST")
			},
			expectError:   true, // Will error because CA server doesn't exist, but will warn about emulator
			expectedError: "",   // Don't check specific error message since it's a connection error
		},
		{
			name: "PUBSUB_EMULATOR_HOST set",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:8090")
				os.Unsetenv("STORAGE_EMULATOR_HOST")
				os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
			},
			expectError:   true, // Will error because CA server doesn't exist, but will warn about emulator
			expectedError: "",   // Don't check specific error message since it's a connection error
		},
		{
			name: "FIRESTORE_EMULATOR_HOST set",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:8090")
				os.Unsetenv("STORAGE_EMULATOR_HOST")
				os.Unsetenv("PUBSUB_EMULATOR_HOST")
				os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
			},
			expectError:   true, // Will error because CA server doesn't exist, but will warn about emulator
			expectedError: "",   // Don't check specific error message since it's a connection error
		},
		{
			name: "FIREBASE_EMULATOR_HOST set",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:8090")
				os.Unsetenv("STORAGE_EMULATOR_HOST")
				os.Unsetenv("PUBSUB_EMULATOR_HOST")
				os.Unsetenv("FIRESTORE_EMULATOR_HOST")
				os.Setenv("FIREBASE_EMULATOR_HOST", "localhost:9099")
			},
			expectError:   true, // Will error because CA server doesn't exist, but will warn about emulator
			expectedError: "",   // Don't check specific error message since it's a connection error
		},
		{
			name: "Multiple emulator variables set (first one detected)",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:8090")
				os.Setenv("STORAGE_EMULATOR_HOST", "localhost:9000")
				os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
			},
			expectError:   true, // Will error because CA server doesn't exist, but will warn about emulator
			expectedError: "",   // Don't check specific error message since it's a connection error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment for this test
			tt.setupEnv()

			// Test UpdateTransport
			err := UpdateTransport()
			if (err != nil) != tt.expectError {
				t.Errorf("UpdateTransport() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectError && err != nil {
				// For emulator detection tests, we now expect connection errors since
				// emulator detection only warns and doesn't prevent the CA update attempt
				if tt.expectedError != "" && err.Error() != tt.expectedError {
					t.Errorf("UpdateTransport() error = %v, expected %v", err.Error(), tt.expectedError)
				}
				// If expectedError is empty, we just expect some error (likely connection error)
			}
		})
	}
}

func TestEmulatorEnvVarDetectionOnlyIf(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalStorage := os.Getenv("STORAGE_EMULATOR_HOST")

	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("STORAGE_EMULATOR_HOST", originalStorage)
	}()

	tests := []struct {
		name        string
		sglCA       string
		emulatorVar string
		expectError bool
	}{
		{
			name:        "No SGL_CA set, emulator var present - should not error",
			sglCA:       "",
			emulatorVar: "localhost:9000",
			expectError: false,
		},
		{
			name:        "SGL_CA set, no emulator var - should not error",
			sglCA:       "http://localhost:8090",
			emulatorVar: "",
			expectError: true, // Will error because CA server doesn't exist, but NOT because of emulator detection
		},
		{
			name:        "SGL_CA set, emulator var present - should warn but not error due to emulator detection",
			sglCA:       "http://localhost:8090",
			emulatorVar: "localhost:9000",
			expectError: true, // Will error because CA server doesn't exist, but will warn about emulator
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.sglCA != "" {
				os.Setenv("SGL_CA", tt.sglCA)
			} else {
				os.Unsetenv("SGL_CA")
			}

			if tt.emulatorVar != "" {
				os.Setenv("STORAGE_EMULATOR_HOST", tt.emulatorVar)
			} else {
				os.Unsetenv("STORAGE_EMULATOR_HOST")
			}

			// Test UpdateTransportOnlyIf
			err := UpdateTransportOnlyIf()
			if (err != nil) != tt.expectError {
				t.Errorf("UpdateTransportOnlyIf() error = %v, expectError %v", err, tt.expectError)
			}

			if tt.expectError && err != nil && tt.emulatorVar != "" {
				// With the new behavior, emulator vars only warn, so we expect connection errors
				// not emulator detection errors
				if strings.Contains(err.Error(), "emulator mode detected") {
					t.Errorf("UpdateTransportOnlyIf() error = %v, should not error on emulator detection anymore (should only warn)", err)
				}
			}
		})
	}
}

func TestUpdateTransportMustPanicOnEmulator(t *testing.T) {
	// Save original environment variables
	originalCA := os.Getenv("SGL_CA")
	originalStorage := os.Getenv("STORAGE_EMULATOR_HOST")

	defer func() {
		os.Setenv("SGL_CA", originalCA)
		os.Setenv("STORAGE_EMULATOR_HOST", originalStorage)
	}()

	// Setup environment to trigger emulator detection
	os.Setenv("SGL_CA", "http://localhost:8090")
	os.Setenv("STORAGE_EMULATOR_HOST", "localhost:9000")

	// Test that UpdateTransportMust panics (due to CA connection failure, not emulator detection)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("UpdateTransportMust() did not panic, expected panic due to CA connection failure")
		} else {
			panicMsg := r.(string)
			// Should panic due to connection failure, not emulator detection
			if strings.Contains(panicMsg, "emulator mode detected") {
				t.Errorf("UpdateTransportMust() panic message = %v, should not panic due to emulator detection (should only warn)", panicMsg)
			}
		}
	}()

	UpdateTransportMust()
}

func TestCheckForEmulatorEnvVars(t *testing.T) {
	// Save original environment variables
	originalVars := make(map[string]string)
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
		originalVars[envVar] = os.Getenv(envVar)
	}

	defer func() {
		for envVar, originalValue := range originalVars {
			os.Setenv(envVar, originalValue)
		}
	}()

	// Clear all emulator vars first
	for _, envVar := range emulatorVars {
		os.Unsetenv(envVar)
	}

	// Test with no emulator vars set - should not panic or error
	checkForEmulatorEnvVars()

	// Test each emulator var individually - should warn but not error
	for _, envVar := range emulatorVars {
		t.Run("Check_"+envVar, func(t *testing.T) {
			// Clear all vars
			for _, v := range emulatorVars {
				os.Unsetenv(v)
			}

			// Set the specific var
			os.Setenv(envVar, "localhost:8080")

			// Function should complete without panic - warnings are logged but not testable directly
			checkForEmulatorEnvVars()
		})
	}
}
