package servicemanager

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	sm := New()
	if sm == nil {
		t.Fatal("Expected ServiceManager instance, got nil")
	}

	// Test default port range
	portRange := sm.GetPortRange()
	if portRange.Start != 80 || portRange.End != 9099 {
		t.Errorf("Expected port range 80-9099, got %d-%d", portRange.Start, portRange.End)
	}

	// Test that known services are initialized
	if len(sm.knownServices) == 0 {
		t.Error("Expected known services to be initialized")
	}

	// Test that monitored ports are initialized
	monitoredPorts := sm.GetMonitoredPorts()
	if len(monitoredPorts) == 0 {
		t.Error("Expected monitored ports to be initialized")
	}
}

func TestNewSimple(t *testing.T) {
	sm := NewSimple()
	if sm == nil {
		t.Fatal("Expected ServiceManager instance, got nil")
	}

	// Test that Docker config is nil for simple manager
	if sm.dockerConfig != nil {
		t.Error("Expected dockerConfig to be nil for simple manager")
	}

	if sm.IsDockerAvailable() {
		t.Error("Expected Docker to be unavailable for simple manager")
	}
}

func TestWithOptions(t *testing.T) {
	// Test with custom port range
	sm := New(
		WithPortRange(3000, 4000),
		WithKnownService(3000, "Test Service", "http://localhost:3000/health", false),
		WithMonitoredPort(3001, "Test Monitor"),
	)

	// Verify port range
	portRange := sm.GetPortRange()
	if portRange.Start != 3000 || portRange.End != 4000 {
		t.Errorf("Expected port range 3000-4000, got %d-%d", portRange.Start, portRange.End)
	}

	// Verify known service
	if config, exists := sm.knownServices[3000]; !exists {
		t.Error("Expected known service on port 3000")
	} else if config.Name != "Test Service" {
		t.Errorf("Expected service name 'Test Service', got '%s'", config.Name)
	}

	// Verify monitored port
	desc := sm.GetPortDescription(3001)
	if desc != "Test Monitor" {
		t.Errorf("Expected description 'Test Monitor', got '%s'", desc)
	}
}

func TestPortManagement(t *testing.T) {
	sm := NewSimple()

	// Test adding monitored port
	sm.AddMonitoredPort(8888, "Test Service")

	monitoredPorts := sm.GetMonitoredPorts()
	found := false
	for _, port := range monitoredPorts {
		if port == 8888 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected port 8888 to be in monitored ports")
	}

	// Test description
	desc := sm.GetPortDescription(8888)
	if desc != "Test Service" {
		t.Errorf("Expected description 'Test Service', got '%s'", desc)
	}

	// Test removing monitored port
	sm.RemoveMonitoredPort(8888)
	monitoredPorts = sm.GetMonitoredPorts()
	for _, port := range monitoredPorts {
		if port == 8888 {
			t.Error("Expected port 8888 to be removed from monitored ports")
		}
	}

	// Test description after removal
	desc = sm.GetPortDescription(8888)
	if desc != "Unknown Service" {
		t.Errorf("Expected description 'Unknown Service', got '%s'", desc)
	}
}

func TestKnownServiceManagement(t *testing.T) {
	sm := NewSimple()

	// Test adding known service
	sm.AddKnownService(9999, "Custom Service", "https://localhost:9999/health", true)

	// Verify it was added
	if config, exists := sm.knownServices[9999]; !exists {
		t.Error("Expected known service on port 9999")
	} else {
		if config.Name != "Custom Service" {
			t.Errorf("Expected name 'Custom Service', got '%s'", config.Name)
		}
		if config.HealthURL != "https://localhost:9999/health" {
			t.Errorf("Expected health URL 'https://localhost:9999/health', got '%s'", config.HealthURL)
		}
		if !config.IsSecure {
			t.Error("Expected service to be marked as secure")
		}
	}
}

func TestPortRangeManagement(t *testing.T) {
	sm := NewSimple()

	// Test setting port range
	sm.SetPortRange(5000, 6000)

	portRange := sm.GetPortRange()
	if portRange.Start != 5000 || portRange.End != 6000 {
		t.Errorf("Expected port range 5000-6000, got %d-%d", portRange.Start, portRange.End)
	}
}

func TestIsPortListening(t *testing.T) {
	sm := NewSimple()

	// Test with a port that should not be listening (high port number)
	if sm.isPortListening(65534) {
		t.Error("Expected port 65534 to not be listening")
	}

	// We can't easily test a listening port without setting one up,
	// so we'll just verify the method doesn't panic
}

func TestSSHProcessDetection(t *testing.T) {
	sm := NewSimple()

	testCases := []struct {
		command  string
		expected bool
	}{
		{"ssh", true},
		{"sshd", true},
		{"ssh-agent", true},
		{"SSH-KEYGEN", true}, // Case insensitive
		{"docker-ssh", true},
		{"nginx", false},
		{"node", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := sm.isSSHProcess(tc.command)
		if result != tc.expected {
			t.Errorf("isSSHProcess(%s) = %v, expected %v", tc.command, result, tc.expected)
		}
	}
}

func TestImagesMatch(t *testing.T) {
	sm := NewSimple()

	testCases := []struct {
		actual   string
		expected string
		matches  bool
	}{
		{"nginx:latest", "nginx:latest", true},
		{"nginx:1.20", "nginx:latest", true}, // Base name matches
		{"nginx", "nginx:latest", true},      // No tag vs with tag
		{"mysql:8.0", "postgres:13", false},  // Different images
		{"", "", true},                       // Both empty
		{"nginx", "", false},                 // One empty
	}

	for _, tc := range testCases {
		result := sm.imagesMatch(tc.actual, tc.expected)
		if result != tc.matches {
			t.Errorf("imagesMatch(%s, %s) = %v, expected %v", tc.actual, tc.expected, result, tc.matches)
		}
	}
}

func TestFormatUptime(t *testing.T) {
	testCases := []struct {
		duration time.Duration
		contains string
	}{
		{30 * time.Second, "seconds"},
		{5 * time.Minute, "minutes"},
		{2 * time.Hour, "hours"},
		{25 * time.Hour, "days"},
	}

	for _, tc := range testCases {
		result := formatUptime(tc.duration)
		if !containsSubstring(result, tc.contains) {
			t.Errorf("formatUptime(%v) = %s, expected to contain '%s'", tc.duration, result, tc.contains)
		}
	}
}

func TestParsePortMapping(t *testing.T) {
	testCases := []struct {
		input            string
		expectedExternal int
		expectedInternal int
		expectedProtocol string
	}{
		{"8080:80", 8080, 80, "tcp"},
		{"8080:80/tcp", 8080, 80, "tcp"},
		{"8080:80/udp", 8080, 80, "udp"},
		{"invalid", 0, 0, "tcp"},
		{"8080", 0, 0, "tcp"},
	}

	for _, tc := range testCases {
		external, internal, protocol := parsePortMapping(tc.input)
		if external != tc.expectedExternal || internal != tc.expectedInternal || protocol != tc.expectedProtocol {
			t.Errorf("parsePortMapping(%s) = (%d, %d, %s), expected (%d, %d, %s)",
				tc.input, external, internal, protocol,
				tc.expectedExternal, tc.expectedInternal, tc.expectedProtocol)
		}
	}
}

func TestGenerateHealthPath(t *testing.T) {
	testCases := []struct {
		serviceName string
		isSecure    bool
		expected    string
	}{
		{"ca", false, "http://localhost/health"},
		{"ca", true, "https://localhost/health"},
		{"firebase", false, "http://localhost/health"},
		{"metadata", false, "http://localhost/"},
		{"metadata", true, "https://localhost/"},
		{"unknown", false, "http://localhost/health"},
		{"unknown", true, "https://localhost/health"},
	}

	for _, tc := range testCases {
		result := generateHealthPath(tc.serviceName, tc.isSecure)
		if result != tc.expected {
			t.Errorf("generateHealthPath(%s, %v) = %s, expected %s",
				tc.serviceName, tc.isSecure, result, tc.expected)
		}
	}
}

func TestParseDependsOn(t *testing.T) {
	testCases := []struct {
		input    interface{}
		expected []string
	}{
		{
			[]interface{}{"service1", "service2"},
			[]string{"service1", "service2"},
		},
		{
			map[string]interface{}{
				"service1": map[string]interface{}{"condition": "service_healthy"},
				"service2": map[string]interface{}{"condition": "service_started"},
			},
			[]string{"service1", "service2"}, // Order may vary, so we check length and contents separately
		},
		{
			nil,
			[]string{},
		},
	}

	for _, tc := range testCases {
		result := parseDependsOn(tc.input)
		if len(result) != len(tc.expected) {
			t.Errorf("parseDependsOn returned %d items, expected %d", len(result), len(tc.expected))
			continue
		}

		// For map inputs, order is not guaranteed, so check each expected item exists
		if len(tc.expected) > 0 {
			expectedMap := make(map[string]bool)
			for _, dep := range tc.expected {
				expectedMap[dep] = true
			}
			for _, dep := range result {
				if !expectedMap[dep] {
					t.Errorf("parseDependsOn returned unexpected dependency: %s", dep)
				}
			}
		}
	}
}

func TestDockerConfig(t *testing.T) {
	sm := New()

	// Test Docker config methods
	dockerConfig := sm.GetDockerConfig()
	if dockerConfig == nil {
		t.Error("Expected Docker config to be non-nil")
	}

	// Test that socket path is set (even if Docker is not available)
	socketPath := sm.GetDockerSocketPath()
	if sm.IsDockerAvailable() && socketPath == "" {
		t.Error("Expected socket path to be set when Docker is available")
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

// Test to ensure version is set
func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Version should be in semantic versioning format (without 'v' prefix)
	if len(Version) < 5 {
		t.Errorf("Version %s should be in semantic version format (e.g., '0.1.0')", Version)
	}
}

// Test ServiceConfig struct
func TestServiceConfig(t *testing.T) {
	config := ServiceConfig{
		Name:      "Test Config",
		HealthURL: "http://localhost:8080/health",
		IsSecure:  false,
	}

	if config.Name != "Test Config" {
		t.Errorf("Expected Name 'Test Config', got '%s'", config.Name)
	}
	if config.HealthURL != "http://localhost:8080/health" {
		t.Errorf("Expected HealthURL 'http://localhost:8080/health', got '%s'", config.HealthURL)
	}
	if config.IsSecure {
		t.Error("Expected IsSecure to be false")
	}
}

// Test PortRange struct
func TestPortRange(t *testing.T) {
	portRange := PortRange{
		Start: 3000,
		End:   4000,
	}

	if portRange.Start != 3000 {
		t.Errorf("Expected Start 3000, got %d", portRange.Start)
	}
	if portRange.End != 4000 {
		t.Errorf("Expected End 4000, got %d", portRange.End)
	}
}
