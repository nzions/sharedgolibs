package basic

import (
	"strings"
	"testing"
)

// TestSimpleAssertions tests basic assertion patterns
// Updated: Testing TUI with FULL test suite - 60 tests!
func TestSimpleAssertions(t *testing.T) {
	if 2+2 != 4 {
		t.Error("Basic math failed")
	}
}

// TestStringOperations tests string manipulation
func TestStringOperations(t *testing.T) {
	input := "Hello, World!"
	expected := "HELLO, WORLD!"
	result := strings.ToUpper(input)

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestSliceOperations tests slice operations
func TestSliceOperations(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}

	if len(slice) != 5 {
		t.Errorf("Expected length 5, got %d", len(slice))
	}

	// Test slice contents
	for i, v := range slice {
		if v != i+1 {
			t.Errorf("Expected %d at index %d, got %d", i+1, i, v)
		}
	}
}

// TestMapOperations tests map operations
func TestMapOperations(t *testing.T) {
	m := make(map[string]int)
	m["one"] = 1
	m["two"] = 2
	m["three"] = 3

	if len(m) != 3 {
		t.Errorf("Expected map length 3, got %d", len(m))
	}

	if m["two"] != 2 {
		t.Errorf("Expected m['two'] = 2, got %d", m["two"])
	}
}

// TestStructOperations tests struct operations
func TestStructOperations(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	p := Person{Name: "Alice", Age: 30}

	if p.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got %q", p.Name)
	}

	if p.Age != 30 {
		t.Errorf("Expected age 30, got %d", p.Age)
	}
}

// TestInterfaceUsage tests interface usage
func TestInterfaceUsage(t *testing.T) {
	type Speaker interface {
		Speak() string
	}

	type Dog struct {
		Name string
	}

	dog := Dog{Name: "Buddy"}
	speak := func(d Dog) string {
		return "Woof! I'm " + d.Name
	}

	result := speak(dog)
	expected := "Woof! I'm Buddy"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestChannelOperations tests basic channel operations
func TestChannelOperations(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42
	result := <-ch

	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// TestErrorHandling tests error handling patterns
func TestErrorHandling(t *testing.T) {
	type CustomError struct {
		Message string
	}

	err := CustomError{Message: "test error"}
	errMsg := err.Message

	if errMsg != "test error" {
		t.Errorf("Expected 'test error', got %q", errMsg)
	}
}
