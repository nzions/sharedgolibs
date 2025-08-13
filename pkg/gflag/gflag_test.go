// SPDX-License-Identifier: CC0-1.0

package gflag

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

func TestFlagSet_String(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		longFlag  string
		shortFlag string
		defValue  string
		expected  string
	}{
		{
			name:      "long flag with equals",
			args:      []string{"--name=test"},
			longFlag:  "name",
			shortFlag: "n",
			defValue:  "default",
			expected:  "test",
		},
		{
			name:      "long flag with space",
			args:      []string{"--name", "test"},
			longFlag:  "name",
			shortFlag: "n",
			defValue:  "default",
			expected:  "test",
		},
		{
			name:      "short flag",
			args:      []string{"-n", "test"},
			longFlag:  "name",
			shortFlag: "n",
			defValue:  "default",
			expected:  "test",
		},
		{
			name:      "default value",
			args:      []string{},
			longFlag:  "name",
			shortFlag: "n",
			defValue:  "default",
			expected:  "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			result := fs.String(tt.longFlag, tt.shortFlag, tt.defValue, "test usage")

			err := fs.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if *result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, *result)
			}
		})
	}
}

func TestFlagSet_Bool(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		longFlag  string
		shortFlag string
		defValue  bool
		expected  bool
	}{
		{
			name:      "long flag true",
			args:      []string{"--verbose"},
			longFlag:  "verbose",
			shortFlag: "v",
			defValue:  false,
			expected:  true,
		},
		{
			name:      "short flag true",
			args:      []string{"-v"},
			longFlag:  "verbose",
			shortFlag: "v",
			defValue:  false,
			expected:  true,
		},
		{
			name:      "long flag explicit true",
			args:      []string{"--verbose=true"},
			longFlag:  "verbose",
			shortFlag: "v",
			defValue:  false,
			expected:  true,
		},
		{
			name:      "long flag explicit false",
			args:      []string{"--verbose=false"},
			longFlag:  "verbose",
			shortFlag: "v",
			defValue:  true,
			expected:  false,
		},
		{
			name:      "default value false",
			args:      []string{},
			longFlag:  "verbose",
			shortFlag: "v",
			defValue:  false,
			expected:  false,
		},
		{
			name:      "default value true",
			args:      []string{},
			longFlag:  "verbose",
			shortFlag: "v",
			defValue:  true,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			result := fs.Bool(tt.longFlag, tt.shortFlag, tt.defValue, "test usage")

			err := fs.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if *result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, *result)
			}
		})
	}
}

func TestFlagSet_Int(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		longFlag  string
		shortFlag string
		defValue  int
		expected  int
	}{
		{
			name:      "long flag with equals",
			args:      []string{"--port=8080"},
			longFlag:  "port",
			shortFlag: "p",
			defValue:  3000,
			expected:  8080,
		},
		{
			name:      "long flag with space",
			args:      []string{"--port", "8080"},
			longFlag:  "port",
			shortFlag: "p",
			defValue:  3000,
			expected:  8080,
		},
		{
			name:      "short flag",
			args:      []string{"-p", "8080"},
			longFlag:  "port",
			shortFlag: "p",
			defValue:  3000,
			expected:  8080,
		},
		{
			name:      "default value",
			args:      []string{},
			longFlag:  "port",
			shortFlag: "p",
			defValue:  3000,
			expected:  3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			result := fs.Int(tt.longFlag, tt.shortFlag, tt.defValue, "test usage")

			err := fs.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if *result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, *result)
			}
		})
	}
}

func TestFlagSet_CombinedShortFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	verbose := fs.Bool("verbose", "v", false, "verbose output")
	debug := fs.Bool("debug", "d", false, "debug output")
	port := fs.Int("port", "p", 3000, "port number")

	tests := []struct {
		name          string
		args          []string
		expectVerbose bool
		expectDebug   bool
		expectPort    int
		expectError   bool
	}{
		{
			name:          "combined bool flags",
			args:          []string{"-vd"},
			expectVerbose: true,
			expectDebug:   true,
			expectPort:    3000,
			expectError:   false,
		},
		{
			name:          "combined with value",
			args:          []string{"-vp", "8080"},
			expectVerbose: true,
			expectDebug:   false,
			expectPort:    8080,
			expectError:   false,
		},
		{
			name:          "combined with inline value",
			args:          []string{"-vp8080"},
			expectVerbose: true,
			expectDebug:   false,
			expectPort:    8080,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			*verbose = false
			*debug = false
			*port = 3000

			err := fs.Parse(tt.args)
			if tt.expectError && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.expectError {
				if *verbose != tt.expectVerbose {
					t.Errorf("verbose: expected %t, got %t", tt.expectVerbose, *verbose)
				}
				if *debug != tt.expectDebug {
					t.Errorf("debug: expected %t, got %t", tt.expectDebug, *debug)
				}
				if *port != tt.expectPort {
					t.Errorf("port: expected %d, got %d", tt.expectPort, *port)
				}
			}
		})
	}
}

func TestFlagSet_Args(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	verbose := fs.Bool("verbose", "v", false, "verbose output")
	name := fs.String("name", "n", "default", "name value")

	tests := []struct {
		name       string
		args       []string
		expectArgs []string
		expectNArg int
		expectArg0 string
	}{
		{
			name:       "flags and args",
			args:       []string{"-v", "--name=test", "arg1", "arg2"},
			expectArgs: []string{"arg1", "arg2"},
			expectNArg: 2,
			expectArg0: "arg1",
		},
		{
			name:       "double dash separator",
			args:       []string{"-v", "--", "--name=test", "arg1"},
			expectArgs: []string{"--name=test", "arg1"},
			expectNArg: 2,
			expectArg0: "--name=test",
		},
		{
			name:       "no args",
			args:       []string{"-v", "--name=test"},
			expectArgs: []string{},
			expectNArg: 0,
			expectArg0: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			*verbose = false
			*name = "default"

			err := fs.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			args := fs.Args()
			if len(args) != len(tt.expectArgs) {
				t.Errorf("args length: expected %d, got %d", len(tt.expectArgs), len(args))
			}
			for i, arg := range args {
				if i < len(tt.expectArgs) && arg != tt.expectArgs[i] {
					t.Errorf("arg[%d]: expected %q, got %q", i, tt.expectArgs[i], arg)
				}
			}

			if fs.NArg() != tt.expectNArg {
				t.Errorf("NArg: expected %d, got %d", tt.expectNArg, fs.NArg())
			}

			if fs.Arg(0) != tt.expectArg0 {
				t.Errorf("Arg(0): expected %q, got %q", tt.expectArg0, fs.Arg(0))
			}
		})
	}
}

func TestErrorHandling_ContinueOnError(t *testing.T) {
	tests := []struct {
		name        string
		setupFlags  func(*FlagSet)
		args        []string
		expectedErr string
	}{
		{
			name: "undefined long flag",
			setupFlags: func(fs *FlagSet) {
				fs.Bool("verbose", "v", false, "verbose output")
			},
			args:        []string{"--undefined"},
			expectedErr: "flag provided but not defined: -undefined",
		},
		{
			name: "undefined short flag",
			setupFlags: func(fs *FlagSet) {
				fs.Bool("verbose", "v", false, "verbose output")
			},
			args:        []string{"-u"},
			expectedErr: "flag provided but not defined: -u",
		},
		{
			name: "missing value for string flag",
			setupFlags: func(fs *FlagSet) {
				fs.String("name", "n", "default", "name value")
			},
			args:        []string{"--name"},
			expectedErr: "flag needs an argument: -name",
		},
		{
			name: "missing value for int flag",
			setupFlags: func(fs *FlagSet) {
				fs.Int("port", "p", 3000, "port number")
			},
			args:        []string{"-p"},
			expectedErr: "flag needs an argument: -p",
		},
		{
			name: "invalid int value",
			setupFlags: func(fs *FlagSet) {
				fs.Int("port", "p", 3000, "port number")
			},
			args:        []string{"-p", "invalid"},
			expectedErr: "strconv.ParseInt: parsing \"invalid\": invalid syntax",
		},
		{
			name: "invalid bool value",
			setupFlags: func(fs *FlagSet) {
				fs.Bool("verbose", "v", false, "verbose output")
			},
			args:        []string{"--verbose=invalid"},
			expectedErr: "strconv.ParseBool: parsing \"invalid\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			tt.setupFlags(fs)

			err := fs.Parse(tt.args)
			if err == nil {
				t.Fatal("expected error but got none")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestErrorHandling_PanicOnError(t *testing.T) {
	tests := []struct {
		name       string
		setupFlags func(*FlagSet)
		args       []string
	}{
		{
			name: "undefined long flag triggers panic",
			setupFlags: func(fs *FlagSet) {
				fs.Bool("verbose", "v", false, "verbose output")
			},
			args: []string{"--undefined"},
		},
		{
			name: "undefined short flag triggers panic",
			setupFlags: func(fs *FlagSet) {
				fs.Bool("verbose", "v", false, "verbose output")
			},
			args: []string{"-u"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", PanicOnError)
			tt.setupFlags(fs)

			defer func() {
				if r := recover(); r == nil {
					t.Fatal("expected panic but didn't get one")
				}
			}()

			fs.Parse(tt.args)
			t.Fatal("expected panic before reaching this point")
		})
	}
}

func TestErrorHandling_UndefinedFlagMessages(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "long flag error message format",
			args:        []string{"--nonexistent"},
			expectedErr: "flag provided but not defined: -nonexistent",
		},
		{
			name:        "short flag error message format",
			args:        []string{"-x"},
			expectedErr: "flag provided but not defined: -x",
		},
		{
			name:        "long flag with value error message format",
			args:        []string{"--missing=value"},
			expectedErr: "flag provided but not defined: -missing",
		},
		{
			name:        "combined short flags with undefined",
			args:        []string{"-xyz"},
			expectedErr: "flag provided but not defined: -x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			// Don't add any flags, so all will be undefined

			err := fs.Parse(tt.args)
			if err == nil {
				t.Fatal("expected error but got none")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestFlagSet_ErrorsLegacy(t *testing.T) {
	tests := []struct {
		name        string
		setupFlags  func(*FlagSet)
		args        []string
		expectError bool
	}{
		{
			name: "undefined long flag",
			setupFlags: func(fs *FlagSet) {
				fs.Bool("verbose", "v", false, "verbose output")
			},
			args:        []string{"--undefined"},
			expectError: true,
		},
		{
			name: "undefined short flag",
			setupFlags: func(fs *FlagSet) {
				fs.Bool("verbose", "v", false, "verbose output")
			},
			args:        []string{"-u"},
			expectError: true,
		},
		{
			name: "missing value for string flag",
			setupFlags: func(fs *FlagSet) {
				fs.String("name", "n", "default", "name value")
			},
			args:        []string{"--name"},
			expectError: true,
		},
		{
			name: "missing value for int flag",
			setupFlags: func(fs *FlagSet) {
				fs.Int("port", "p", 3000, "port number")
			},
			args:        []string{"-p"},
			expectError: true,
		},
		{
			name: "invalid int value",
			setupFlags: func(fs *FlagSet) {
				fs.Int("port", "p", 3000, "port number")
			},
			args:        []string{"-p", "invalid"},
			expectError: true,
		},
		{
			name: "invalid bool value",
			setupFlags: func(fs *FlagSet) {
				fs.Bool("verbose", "v", false, "verbose output")
			},
			args:        []string{"--verbose=invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			tt.setupFlags(fs)

			err := fs.Parse(tt.args)
			if tt.expectError && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	testErr := fmt.Errorf("test error")

	tests := []struct {
		name          string
		errorHandling ErrorHandling
		expectPanic   bool
		expectReturn  bool
	}{
		{
			name:          "ContinueOnError returns error",
			errorHandling: ContinueOnError,
			expectPanic:   false,
			expectReturn:  true,
		},
		{
			name:          "PanicOnError panics",
			errorHandling: PanicOnError,
			expectPanic:   true,
			expectReturn:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", tt.errorHandling)

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Fatal("expected panic but didn't get one")
					}
				}()
			}

			err := fs.handleError(testErr)

			if tt.expectReturn && err == nil {
				t.Fatal("expected error to be returned but got nil")
			}
			if tt.expectReturn && err.Error() != testErr.Error() {
				t.Errorf("expected error %q, got %q", testErr.Error(), err.Error())
			}

			if tt.expectPanic {
				t.Fatal("expected panic before reaching this point")
			}
		})
	}
}

func TestUndefinedFlagInCombinedShortFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.Bool("verbose", "v", false, "verbose output")
	fs.Bool("debug", "d", false, "debug output")
	// Note: no 'x' flag defined

	err := fs.Parse([]string{"-vdx"})
	if err == nil {
		t.Fatal("expected error for undefined flag in combined short flags")
	}

	expectedErr := "flag provided but not defined: -x"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestErrorHandlingWithPackageLevelParse(t *testing.T) {
	// Test that package-level functions handle errors correctly
	// Since CommandLine uses ExitOnError, we can't test it directly
	// but we can verify the mechanism works by testing with a custom flagset

	originalCommandLine := CommandLine
	defer func() { CommandLine = originalCommandLine }()

	// Replace CommandLine with a test flagset that uses ContinueOnError
	CommandLine = NewFlagSet("test", ContinueOnError)
	CommandLine.Bool("test-flag", "t", false, "test flag")

	// Test that undefined flags are caught in package-level parsing
	// We can't call Parse() directly as it would use os.Args, so we test
	// the underlying mechanism by calling Parse on our test CommandLine
	err := CommandLine.Parse([]string{"--undefined"})
	if err == nil {
		t.Fatal("expected error for undefined flag")
	}

	expectedErr := "flag provided but not defined: -undefined"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestPackageFunctions(t *testing.T) {
	// Test that package-level functions work
	// Note: We can't easily test the CommandLine functions in unit tests
	// because they share global state, but we can at least verify they exist

	// Create a separate flagset to test the same logic
	fs := NewFlagSet("test", ContinueOnError)

	str := fs.String("string", "", "default", "string flag")
	boolean := fs.Bool("bool", "", false, "bool flag")
	integer := fs.Int("int", "", 0, "int flag")

	err := fs.Parse([]string{"--string=test", "--bool", "--int=42"})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if *str != "test" {
		t.Errorf("string flag: expected 'test', got '%s'", *str)
	}
	if !*boolean {
		t.Errorf("bool flag: expected true, got %t", *boolean)
	}
	if *integer != 42 {
		t.Errorf("int flag: expected 42, got %d", *integer)
	}
}

func TestEdgeCases(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	verbose := fs.Bool("verbose", "v", false, "verbose output")

	t.Run("empty flag string", func(t *testing.T) {
		*verbose = false
		err := fs.Parse([]string{""})
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		// Empty string should be treated as a non-flag argument
		if len(fs.Args()) != 1 || fs.Args()[0] != "" {
			t.Errorf("expected empty string as argument, got %v", fs.Args())
		}
	})

	t.Run("single dash", func(t *testing.T) {
		*verbose = false
		err := fs.Parse([]string{"-"})
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		// Single dash should be treated as a non-flag argument
		if len(fs.Args()) != 1 || fs.Args()[0] != "-" {
			t.Errorf("expected single dash as argument, got %v", fs.Args())
		}
	})

	t.Run("arg bounds", func(t *testing.T) {
		*verbose = false
		err := fs.Parse([]string{"arg1", "arg2"})
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if fs.Arg(-1) != "" {
			t.Errorf("Arg(-1) should return empty string")
		}
		if fs.Arg(10) != "" {
			t.Errorf("Arg(10) should return empty string")
		}
	})
}

// Benchmark tests
func BenchmarkFlagSet_Parse(b *testing.B) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.Bool("verbose", "v", false, "verbose output")
	fs.String("name", "n", "default", "name value")
	fs.Int("port", "p", 3000, "port number")

	args := []string{"-v", "--name=test", "-p", "8080", "arg1", "arg2"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Parse(args)
	}
}

func BenchmarkFlagSet_CombinedFlags(b *testing.B) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.Bool("verbose", "v", false, "verbose output")
	fs.Bool("debug", "d", false, "debug output")
	fs.Bool("quiet", "q", false, "quiet output")

	args := []string{"-vdq"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Parse(args)
	}
}

func TestAutoHelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "help flag automatically added",
			args: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)

			// Check that help flag was automatically added
			helpFlag := fs.GetFlag("help")
			if helpFlag == nil {
				t.Error("help flag should be automatically added")
			}

			if helpFlag.Name != "help" {
				t.Errorf("expected help flag name to be 'help', got %q", helpFlag.Name)
			}

			if helpFlag.ShortName != "h" {
				t.Errorf("expected help flag short name to be 'h', got %q", helpFlag.ShortName)
			}

			if helpFlag.Usage != "show help message" {
				t.Errorf("expected help flag usage to be 'show help message', got %q", helpFlag.Usage)
			}
		})
	}
}

func TestHelpFlagNotAddedIfExists(t *testing.T) {
	// Create a flagset and manually add help flag before calling NewFlagSet's automatic addition
	fs := &FlagSet{
		name:          "test",
		flags:         make(map[string]*Flag),
		shortMap:      make(map[string]*Flag),
		errorHandling: ContinueOnError,
	}
	fs.usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", fs.name)
		fs.PrintDefaults()
	}

	// Add a custom help flag before the automatic one would be added
	fs.BoolVar(new(bool), "help", "", false, "custom help message")

	// Now call the method that would add automatic help
	fs.addHelpFlagIfNotExists()

	// Check that our custom help flag is preserved
	helpFlag := fs.GetFlag("help")
	if helpFlag == nil {
		t.Error("help flag should exist")
	}

	if helpFlag.ShortName != "" {
		t.Errorf("expected custom help flag short name to be empty, got %q", helpFlag.ShortName)
	}

	if helpFlag.Usage != "custom help message" {
		t.Errorf("expected custom help flag usage to be 'custom help message', got %q", helpFlag.Usage)
	}
}

func TestHelpFlagTriggersUsage(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.AddString("name", "n", "default", "a name flag")

	// Test --help flag
	err := fs.Parse([]string{"--help"})
	if err != nil {
		t.Errorf("help flag should not return error with ContinueOnError, got: %v", err)
	}

	// Test -h flag
	fs2 := NewFlagSet("test", ContinueOnError)
	fs2.AddString("name", "n", "default", "a name flag")

	err = fs2.Parse([]string{"-h"})
	if err != nil {
		t.Errorf("help flag should not return error with ContinueOnError, got: %v", err)
	}
}

// TestExitOnError_Subprocess tests the ExitOnError behavior using subprocess
func TestExitOnError_Subprocess(t *testing.T) {
	// Path to the test program
	testProgram := "testdata/exit_test_program.go"

	// Build the test program first
	tempBinary := "testdata/exit_test_program_binary"
	buildCmd := exec.Command("go", "build", "-o", tempBinary, testProgram)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build test program: %v", err)
	}
	defer os.Remove(tempBinary) // Clean up

	tests := []struct {
		name         string
		args         []string
		expectedCode int
		expectStderr string
	}{
		{
			name:         "undefined long flag exits with code 2",
			args:         []string{"--undefined"},
			expectedCode: 2,
			expectStderr: "flag provided but not defined: -undefined",
		},
		{
			name:         "undefined short flag exits with code 2",
			args:         []string{"-u"},
			expectedCode: 2,
			expectStderr: "flag provided but not defined: -u",
		},
		{
			name:         "help flag exits with code 0",
			args:         []string{"--help"},
			expectedCode: 0,
			expectStderr: "Usage of",
		},
		{
			name:         "valid flag does not exit",
			args:         []string{"--valid"},
			expectedCode: 0,
			expectStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the test program with specific arguments
			cmd := exec.Command("./" + tempBinary)
			cmd.Args = append(cmd.Args, tt.args...)

			output, err := cmd.CombinedOutput()

			// Check exit code
			var exitCode int
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				exitCode = 0
			}

			if exitCode != tt.expectedCode {
				t.Errorf("expected exit code %d, got %d. Output: %s", tt.expectedCode, exitCode, string(output))
			}

			// Check stderr output if expected
			if tt.expectStderr != "" {
				outputStr := string(output)
				if !strings.Contains(outputStr, tt.expectStderr) {
					t.Errorf("expected output to contain %q, got %q", tt.expectStderr, outputStr)
				}
			}
		})
	}
}
