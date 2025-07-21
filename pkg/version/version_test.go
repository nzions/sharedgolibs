package version

import "testing"

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("Version should not be empty")
	}
	
	if version != Version {
		t.Errorf("GetVersion() = %q, want %q", version, Version)
	}
	
	// Version should follow semantic versioning pattern
	if len(version) < 5 || version[0] != 'v' {
		t.Errorf("Version %q should follow vX.Y.Z format", version)
	}
}
