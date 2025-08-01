package failures

import (
	"context"
	"testing"
	"time"
)

// TestLongRunning simulates a test that takes a very long time
func TestLongRunning(t *testing.T) {
	// Simulate long-running operation
	time.Sleep(5 * time.Second)
	t.Log("Long running test completed")
}

// TestHanging simulates a test that hangs indefinitely
func TestHanging(t *testing.T) {
	// Create a channel and wait forever
	ch := make(chan bool)
	t.Log("Starting hanging test...")
	<-ch // This will block forever
	t.Log("This should never be reached")
}

// TestDeadlock simulates a deadlock condition
func TestDeadlock(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)

	go func() {
		ch1 <- 1
		<-ch2
	}()

	go func() {
		ch2 <- 2
		<-ch1
	}()

	// Wait for both goroutines (will deadlock)
	time.Sleep(100 * time.Millisecond)
	t.Log("This test will deadlock")
}

// TestSlowWithContext demonstrates a slow test that respects context
func TestSlowWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case <-time.After(5 * time.Second):
		t.Log("Operation completed")
	case <-ctx.Done():
		t.Error("Test timed out as expected")
	}
}

// TestBusyLoop simulates a CPU-intensive infinite loop
func TestBusyLoop(t *testing.T) {
	t.Log("Starting busy loop...")

	// Busy loop that consumes CPU
	counter := 0
	for {
		counter++
		if counter%1000000 == 0 {
			// Check if we should stop (in real scenarios, this might check context)
			if counter > 100000000 { // Arbitrary large number
				break
			}
		}
	}

	t.Logf("Busy loop completed with counter: %d", counter)
}

// TestMemoryLeak simulates a test that leaks memory
func TestMemoryLeak(t *testing.T) {
	// Allocate increasingly large slices
	var leaks [][]byte

	for i := 0; i < 1000; i++ {
		// Allocate 1MB chunks
		chunk := make([]byte, 1024*1024)
		leaks = append(leaks, chunk)

		if i%100 == 0 {
			t.Logf("Allocated %d MB", i)
		}
	}

	t.Logf("Memory leak test completed, allocated %d chunks", len(leaks))
}

// TestSlowNetwork simulates a slow network operation
func TestSlowNetwork(t *testing.T) {
	// Simulate network delay
	t.Log("Simulating slow network operation...")

	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		t.Logf("Network operation %d/10 completed", i+1)
	}

	t.Log("Slow network test completed")
}

// TestTimeoutWithCleanup demonstrates proper cleanup on timeout
func TestTimeoutWithCleanup(t *testing.T) {
	// Set up some resource
	resource := "important resource"
	t.Logf("Setting up %s", resource)

	defer func() {
		t.Logf("Cleaning up %s", resource)
	}()

	// Simulate long operation
	time.Sleep(3 * time.Second)

	t.Log("Test completed successfully")
}

// TestGoroutineLeak demonstrates goroutine leak
func TestGoroutineLeak(t *testing.T) {
	// Start many goroutines that don't finish
	for i := 0; i < 100; i++ {
		go func(id int) {
			// Infinite loop
			for {
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	time.Sleep(1 * time.Second)
	t.Log("Goroutine leak test completed")
}
