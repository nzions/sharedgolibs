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
			expectError:   true,
			expectedError: "emulator mode detected: transport update not allowed when Google Cloud emulators are active: STORAGE_EMULATOR_HOST=localhost:9000",
		},
		{
			name: "PUBSUB_EMULATOR_HOST set",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:8090")
				os.Unsetenv("STORAGE_EMULATOR_HOST")
				os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
			},
			expectError:   true,
			expectedError: "emulator mode detected: transport update not allowed when Google Cloud emulators are active: PUBSUB_EMULATOR_HOST=localhost:8085",
		},
		{
			name: "FIRESTORE_EMULATOR_HOST set",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:8090")
				os.Unsetenv("STORAGE_EMULATOR_HOST")
				os.Unsetenv("PUBSUB_EMULATOR_HOST")
				os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
			},
			expectError:   true,
			expectedError: "emulator mode detected: transport update not allowed when Google Cloud emulators are active: FIRESTORE_EMULATOR_HOST=localhost:8080",
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
			expectError:   true,
			expectedError: "emulator mode detected: transport update not allowed when Google Cloud emulators are active: FIREBASE_EMULATOR_HOST=localhost:9099",
		},
		{
			name: "Multiple emulator variables set (first one detected)",
			setupEnv: func() {
				os.Setenv("SGL_CA", "http://localhost:8090")
				os.Setenv("STORAGE_EMULATOR_HOST", "localhost:9000")
				os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
			},
			expectError:   true,
			expectedError: "emulator mode detected: transport update not allowed when Google Cloud emulators are active: STORAGE_EMULATOR_HOST=localhost:9000",
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
				if strings.Contains(err.Error(), "emulator mode detected") {
					// This is the emulator detection error we're testing for
					if tt.expectedError != "" && err.Error() != tt.expectedError {
						t.Errorf("UpdateTransport() error = %v, expected %v", err.Error(), tt.expectedError)
					}
				} else if tt.expectedError != "" {
					// We expected an emulator error but got a different error
					t.Errorf("UpdateTransport() error = %v, expected emulator detection error", err)
				}
				// Otherwise it's probably a CA server connection error which is expected for the "no emulator" test
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
			name:        "SGL_CA set, emulator var present - should error",
			sglCA:       "http://localhost:8090",
			emulatorVar: "localhost:9000",
			expectError: true,
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
				// Only check for emulator detection if we set an emulator var
				if !strings.Contains(err.Error(), "emulator mode detected") {
					t.Errorf("UpdateTransportOnlyIf() error = %v, expected error to contain 'emulator mode detected'", err)
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

	// Test that UpdateTransportMust panics
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("UpdateTransportMust() did not panic, expected panic due to emulator detection")
		} else {
			panicMsg := r.(string)
			if !strings.Contains(panicMsg, "emulator mode detected") {
				t.Errorf("UpdateTransportMust() panic message = %v, expected to contain 'emulator mode detected'", panicMsg)
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

	// Test with no emulator vars set
	err := checkForEmulatorEnvVars()
	if err != nil {
		t.Errorf("checkForEmulatorEnvVars() with no vars set should not error, got: %v", err)
	}

	// Test each emulator var individually
	for _, envVar := range emulatorVars {
		t.Run("Check_"+envVar, func(t *testing.T) {
			// Clear all vars
			for _, v := range emulatorVars {
				os.Unsetenv(v)
			}

			// Set the specific var
			os.Setenv(envVar, "localhost:8080")

			err := checkForEmulatorEnvVars()
			if err == nil {
				t.Errorf("checkForEmulatorEnvVars() with %s set should error", envVar)
			} else if !strings.Contains(err.Error(), envVar) {
				t.Errorf("checkForEmulatorEnvVars() error should mention %s, got: %v", envVar, err)
			}
		})
	}
}
