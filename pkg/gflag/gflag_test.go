// SPDX-License-Identifier: CC0-1.0

package gflag

import (
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

func TestFlagSet_Errors(t *testing.T) {
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
