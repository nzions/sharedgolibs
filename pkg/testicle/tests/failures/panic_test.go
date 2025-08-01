package failures

import (
	"testing"
)

// TestPanicRecovery demonstrates a test that panics
func TestPanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test panicked: %v", r)
		}
	}()

	// This will cause a panic
	panic("Intentional panic for testing")
}

// TestPanicInGoroutine demonstrates panic in a goroutine
func TestPanicInGoroutine(t *testing.T) {
	done := make(chan bool)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Goroutine panicked: %v", r)
			}
			done <- true
		}()

		panic("Panic in goroutine")
	}()

	<-done
}

// TestNilPointerPanic demonstrates nil pointer panic
func TestNilPointerPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Nil pointer panic: %v", r)
		}
	}()

	var slice []int
	// This will panic with index out of range
	_ = slice[0]
}

// TestMapPanic demonstrates map panic
func TestMapPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Map panic: %v", r)
		}
	}()

	var m map[string]int
	// This will panic - writing to nil map
	m["key"] = 42
}

// TestDivisionByZeroPanic demonstrates division by zero
func TestDivisionByZeroPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Division by zero panic: %v", r)
		}
	}()

	zero := 0
	result := 42 / zero
	t.Logf("Result: %d", result) // Should never reach here
}

// TestChannelPanic demonstrates channel panic
func TestChannelPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Channel panic: %v", r)
		}
	}()

	ch := make(chan int)
	close(ch)
	// This will panic - sending on closed channel
	ch <- 42
}

// TestRecursionPanic demonstrates stack overflow from infinite recursion
func TestRecursionPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Recursion panic: %v", r)
		}
	}()

	var recursiveFunc func()
	recursiveFunc = func() {
		recursiveFunc() // Infinite recursion
	}

	recursiveFunc()
}

// TestSliceBoundsPanic demonstrates slice bounds panic
func TestSliceBoundsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Slice bounds panic: %v", r)
		}
	}()

	slice := []int{1, 2, 3}
	// This will panic - index out of range
	_ = slice[10]
}

// TestInterfacePanic demonstrates interface panic
func TestInterfacePanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Interface panic: %v", r)
		}
	}()

	var i interface{}
	// This will panic - type assertion on nil interface
	_ = i.(string)
}

// TestUnrecoveredPanic demonstrates a panic that is not recovered
func TestUnrecoveredPanic(t *testing.T) {
	// No defer recover() - this panic will terminate the test
	panic("Unrecovered panic - this will terminate the test")
}
