package subtests

import (
	"fmt"
	"strings"
	"testing"
)

// TestTableDriven demonstrates table-driven tests with subtests
func TestTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Single word", "hello", "HELLO"},
		{"Multiple words", "hello world", "HELLO WORLD"},
		{"With numbers", "hello123", "HELLO123"},
		{"With special chars", "hello-world_test", "HELLO-WORLD_TEST"},
		{"Unicode", "héllo wörld", "HÉLLO WÖRLD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.ToUpper(tt.input)
			if result != tt.expected {
				t.Errorf("ToUpper(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNestedSubtests demonstrates nested subtests
func TestNestedSubtests(t *testing.T) {
	t.Run("StringOperations", func(t *testing.T) {
		t.Run("ToUpper", func(t *testing.T) {
			input := "hello"
			expected := "HELLO"
			result := strings.ToUpper(input)
			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
		})

		t.Run("ToLower", func(t *testing.T) {
			input := "HELLO"
			expected := "hello"
			result := strings.ToLower(input)
			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
		})

		t.Run("Contains", func(t *testing.T) {
			tests := []struct {
				haystack string
				needle   string
				expected bool
			}{
				{"hello world", "world", true},
				{"hello world", "universe", false},
				{"", "", true},
				{"hello", "", true},
			}

			for _, tt := range tests {
				t.Run(fmt.Sprintf("%s_contains_%s", tt.haystack, tt.needle), func(t *testing.T) {
					result := strings.Contains(tt.haystack, tt.needle)
					if result != tt.expected {
						t.Errorf("Contains(%q, %q) = %v, want %v", tt.haystack, tt.needle, result, tt.expected)
					}
				})
			}
		})
	})

	t.Run("NumberOperations", func(t *testing.T) {
		t.Run("Addition", func(t *testing.T) {
			tests := []struct {
				a, b, expected int
			}{
				{1, 2, 3},
				{0, 0, 0},
				{-1, 1, 0},
				{10, -5, 5},
			}

			for _, tt := range tests {
				t.Run(fmt.Sprintf("%d_plus_%d", tt.a, tt.b), func(t *testing.T) {
					result := tt.a + tt.b
					if result != tt.expected {
						t.Errorf("%d + %d = %d, want %d", tt.a, tt.b, result, tt.expected)
					}
				})
			}
		})

		t.Run("Multiplication", func(t *testing.T) {
			tests := []struct {
				a, b, expected int
			}{
				{2, 3, 6},
				{0, 5, 0},
				{-2, 3, -6},
				{-2, -3, 6},
			}

			for _, tt := range tests {
				t.Run(fmt.Sprintf("%d_times_%d", tt.a, tt.b), func(t *testing.T) {
					result := tt.a * tt.b
					if result != tt.expected {
						t.Errorf("%d * %d = %d, want %d", tt.a, tt.b, result, tt.expected)
					}
				})
			}
		})
	})
}

// TestComplexStructure demonstrates complex nested subtest structure
func TestComplexStructure(t *testing.T) {
	type User struct {
		ID   int
		Name string
		Age  int
	}

	users := []User{
		{1, "Alice", 30},
		{2, "Bob", 25},
		{3, "Charlie", 35},
	}

	for _, user := range users {
		t.Run(fmt.Sprintf("User_%d_%s", user.ID, user.Name), func(t *testing.T) {
			t.Run("ValidateID", func(t *testing.T) {
				if user.ID <= 0 {
					t.Errorf("Invalid user ID: %d", user.ID)
				}
			})

			t.Run("ValidateName", func(t *testing.T) {
				if user.Name == "" {
					t.Error("User name cannot be empty")
				}
				if len(user.Name) < 2 {
					t.Errorf("User name too short: %q", user.Name)
				}
			})

			t.Run("ValidateAge", func(t *testing.T) {
				if user.Age < 0 {
					t.Errorf("Invalid age: %d", user.Age)
				}
				if user.Age > 120 {
					t.Errorf("Age seems unrealistic: %d", user.Age)
				}
			})

			t.Run("BusinessLogic", func(t *testing.T) {
				t.Run("CanVote", func(t *testing.T) {
					canVote := user.Age >= 18
					if user.Age >= 18 && !canVote {
						t.Error("User should be able to vote")
					}
					if user.Age < 18 && canVote {
						t.Error("User should not be able to vote")
					}
				})

				t.Run("IsAdult", func(t *testing.T) {
					isAdult := user.Age >= 21
					if user.Age >= 21 && !isAdult {
						t.Error("User should be considered adult")
					}
				})
			})
		})
	}
}

// TestDynamicSubtests demonstrates dynamically created subtests
func TestDynamicSubtests(t *testing.T) {
	// Generate test cases dynamically
	for i := 1; i <= 10; i++ {
		t.Run(fmt.Sprintf("Square_%d", i), func(t *testing.T) {
			expected := i * i
			result := square(i)
			if result != expected {
				t.Errorf("square(%d) = %d, want %d", i, result, expected)
			}
		})
	}

	// Test different data sizes
	sizes := []int{10, 100, 1000}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("ProcessSlice_Size_%d", size), func(t *testing.T) {
			slice := make([]int, size)
			for i := range slice {
				slice[i] = i
			}

			// Simple processing function
			sum := 0
			for _, v := range slice {
				sum += v
			}

			expected := size * (size - 1) / 2 // Sum of 0 to n-1
			if sum != expected {
				t.Errorf("Sum of slice size %d = %d, want %d", size, sum, expected)
			}
		})
	}
}

// Helper function for testing
func square(n int) int {
	return n * n
}

// TestParallelSubtests demonstrates parallel subtests
func TestParallelSubtests(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"Small", 100},
		{"Medium", 1000},
		{"Large", 10000},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // This subtest can run in parallel

			// Simulate some work
			slice := make([]int, tt.size)
			for i := range slice {
				slice[i] = i * i
			}

			// Verify the slice
			if len(slice) != tt.size {
				t.Errorf("Expected slice length %d, got %d", tt.size, len(slice))
			}
		})
	}
}
