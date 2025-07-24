package binarycleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	config := Config{
		Directory: "/tmp",
		DryRun:    true,
		Verbose:   false,
		Recursive: true,
	}

	bc := New(config)
	if bc == nil {
		t.Fatal("New() returned nil")
	}

	if bc.config != config {
		t.Error("Config not properly set")
	}
}

func TestBinaryType_String(t *testing.T) {
	tests := []struct {
		bt       BinaryType
		expected string
	}{
		{Unknown, "Unknown"},
		{MachO, "Mach-O"},
		{ELF, "ELF"},
	}

	for _, test := range tests {
		if got := test.bt.String(); got != test.expected {
			t.Errorf("BinaryType.String() = %q, want %q", got, test.expected)
		}
	}
}

func TestDetectBinaryType(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "binarycleaner_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		content  []byte
		expected BinaryType
	}{
		{
			name:     "elf_file",
			content:  []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00},
			expected: ELF,
		},
		{
			name:     "macho_32_le",
			content:  []byte{0xfe, 0xed, 0xfa, 0xce, 0x00, 0x00, 0x00, 0x00},
			expected: MachO,
		},
		{
			name:     "macho_64_le",
			content:  []byte{0xfe, 0xed, 0xfa, 0xcf, 0x00, 0x00, 0x00, 0x00},
			expected: MachO,
		},
		{
			name:     "macho_universal",
			content:  []byte{0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00, 0x00},
			expected: MachO,
		},
		{
			name:     "text_file",
			content:  []byte("Hello, World!"),
			expected: Unknown,
		},
		{
			name:     "short_file",
			content:  []byte{0x01, 0x02},
			expected: Unknown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, test.name)
			err := os.WriteFile(filePath, test.content, 0755)
			if err != nil {
				t.Fatal(err)
			}

			got, err := detectBinaryType(filePath)
			if err != nil {
				t.Fatalf("detectBinaryType() error = %v", err)
			}

			if got != test.expected {
				t.Errorf("detectBinaryType() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestIsExecutable(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "binarycleaner_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create executable file
	execFile := filepath.Join(tmpDir, "executable")
	err = os.WriteFile(execFile, []byte("test"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create non-executable file
	nonExecFile := filepath.Join(tmpDir, "nonexecutable")
	err = os.WriteFile(nonExecFile, []byte("test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if !isExecutable(execFile) {
		t.Error("Expected executable file to be detected as executable")
	}

	if isExecutable(nonExecFile) {
		t.Error("Expected non-executable file to be detected as non-executable")
	}

	// Test non-existent file
	if isExecutable(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("Expected non-existent file to be detected as non-executable")
	}
}

func TestFindBinaries(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "binarycleaner_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []struct {
		name     string
		content  []byte
		mode     os.FileMode
		isBinary bool
	}{
		{"elf_binary", []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}, 0755, true},
		{"macho_binary", []byte{0xfe, 0xed, 0xfa, 0xce, 0x00, 0x00, 0x00, 0x00}, 0755, true},
		{"text_file", []byte("Hello, World!"), 0644, false},
		{"non_exec_binary", []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}, 0644, false},
		{"go_file.go", []byte("package main"), 0755, false},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.name)
		err := os.WriteFile(filePath, tf.content, tf.mode)
		if err != nil {
			t.Fatal(err)
		}
	}

	config := Config{
		Directory: tmpDir,
		DryRun:    true,
		Verbose:   false,
		Recursive: false,
	}

	bc := New(config)
	binaries, err := bc.FindBinaries()
	if err != nil {
		t.Fatalf("FindBinaries() error = %v", err)
	}

	expectedCount := 2 // elf_binary and macho_binary
	if len(binaries) != expectedCount {
		t.Errorf("FindBinaries() found %d binaries, want %d", len(binaries), expectedCount)
	}

	// Check that we found the right binaries
	foundTypes := make(map[BinaryType]bool)
	for _, binary := range binaries {
		foundTypes[binary.Type] = true
	}

	if !foundTypes[ELF] || !foundTypes[MachO] {
		t.Error("Expected to find both ELF and Mach-O binaries")
	}
}

func TestClean_DryRun(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "binarycleaner_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test binary
	binaryPath := filepath.Join(tmpDir, "test_binary")
	err = os.WriteFile(binaryPath, []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	config := Config{
		Directory: tmpDir,
		DryRun:    true,
		Verbose:   false,
		Recursive: false,
	}

	bc := New(config)
	err = bc.Clean()
	if err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	// File should still exist in dry run mode
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Error("File was removed in dry run mode")
	}
}

func TestClean_ActualRemoval(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "binarycleaner_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test binary
	binaryPath := filepath.Join(tmpDir, "test_binary")
	err = os.WriteFile(binaryPath, []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	config := Config{
		Directory: tmpDir,
		DryRun:    false,
		Verbose:   false,
		Recursive: false,
	}

	bc := New(config)
	err = bc.Clean()
	if err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	// File should be removed
	if _, err := os.Stat(binaryPath); !os.IsNotExist(err) {
		t.Error("File was not removed")
	}
}
