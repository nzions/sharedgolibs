package failures

import (
	"testing"
)

// TestAssertionFailure demonstrates a basic assertion failure
func TestAssertionFailure(t *testing.T) {
	expected := 5
	actual := 3

	if actual != expected {
		t.Errorf("Intentional failure: expected %d, got %d", expected, actual)
	}
}

// TestMultipleFailures demonstrates multiple assertion failures in one test
func TestMultipleFailures(t *testing.T) {
	t.Error("First failure message")
	t.Error("Second failure message")
	t.Error("Third failure message")
}

// TestStringMismatch demonstrates string comparison failure
func TestStringMismatch(t *testing.T) {
	expected := "Hello, World!"
	actual := "Hello, Universe!"

	if actual != expected {
		t.Errorf("String mismatch:\nExpected: %q\nActual:   %q", expected, actual)
	}
}

// TestSliceMismatch demonstrates slice comparison failure
func TestSliceMismatch(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	actual := []int{1, 2, 3, 4, 6}

	if len(expected) != len(actual) {
		t.Errorf("Length mismatch: expected %d, got %d", len(expected), len(actual))
	}

	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("Element %d mismatch: expected %d, got %d", i, expected[i], actual[i])
		}
	}
}

// TestMapMismatch demonstrates map comparison failure
func TestMapMismatch(t *testing.T) {
	expected := map[string]int{"a": 1, "b": 2, "c": 3}
	actual := map[string]int{"a": 1, "b": 2, "c": 4}

	for key, expectedVal := range expected {
		if actualVal, exists := actual[key]; !exists {
			t.Errorf("Key %q missing from actual map", key)
		} else if actualVal != expectedVal {
			t.Errorf("Value mismatch for key %q: expected %d, got %d", key, expectedVal, actualVal)
		}
	}
}

// TestNilPointerAccess demonstrates nil pointer access
func TestNilPointerAccess(t *testing.T) {
	var ptr *string

	// This should fail safely
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test failed with panic: %v", r)
		}
	}()

	// Intentionally access nil pointer
	if ptr != nil {
		_ = *ptr
	} else {
		t.Error("Pointer is nil as expected, marking as failure for testing")
	}
}

// TestFatal demonstrates fatal error that stops test execution
func TestFatal(t *testing.T) {
	t.Fatal("This is a fatal error - test execution stops here")
	t.Error("This line should never be reached")
}

// TestFatalf demonstrates formatted fatal error
func TestFatalf(t *testing.T) {
	value := 42
	t.Fatalf("Fatal error with value: %d", value)
	t.Error("This line should never be reached")
}

// TestFailNow demonstrates immediate test termination
func TestFailNow(t *testing.T) {
	t.Error("This error is logged")
	t.FailNow()
	t.Error("This error should never be logged")
}

// TestComplexFailure demonstrates a more complex failure scenario
func TestComplexFailure(t *testing.T) {
	type User struct {
		ID   int
		Name string
		Age  int
	}

	expected := User{ID: 1, Name: "Alice", Age: 30}
	actual := User{ID: 1, Name: "Bob", Age: 25}

	if actual.ID != expected.ID {
		t.Errorf("ID mismatch: expected %d, got %d", expected.ID, actual.ID)
	}

	if actual.Name != expected.Name {
		t.Errorf("Name mismatch: expected %q, got %q", expected.Name, actual.Name)
	}

	if actual.Age != expected.Age {
		t.Errorf("Age mismatch: expected %d, got %d", expected.Age, actual.Age)
	}
}
