//go:build linux

// SPDX-License-Identifier: CC0-1.0

package waitlib

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// setProcessTitle sets the process title on Linux using multiple approaches
func setProcessTitle(title string) error {
	// Method 1: Use prctl(PR_SET_NAME) syscall for the comm field (limited to 15 chars)
	if err := setProcessNamePrctl(title); err == nil {
		// Also try to update the longer argv[0] for ps output
		setProcessNameArgv(title)
		return nil
	}

	// Method 2: Fallback to /proc/self/comm if prctl fails
	if _, err := os.Stat("/proc/self/comm"); err == nil {
		// Truncate to 15 characters (Linux kernel limit for comm)
		truncated := title
		if len(truncated) > 15 {
			truncated = title[:15]
		}
		if err := os.WriteFile("/proc/self/comm", []byte(truncated), 0644); err == nil {
			// Also try to update argv[0]
			setProcessNameArgv(title)
			return nil
		}
	}

	// Method 3: Just try argv[0] manipulation as last resort
	return setProcessNameArgv(title)
}

// setProcessNamePrctl uses the prctl syscall to set the process name (Linux only)
func setProcessNamePrctl(title string) error {
	// PR_SET_NAME = 15, SYS_PRCTL = 157 on most Linux systems
	const PR_SET_NAME = 15
	const SYS_PRCTL = 157

	// Truncate to 15 characters (kernel limit)
	name := title
	if len(name) > 15 {
		name = title[:15]
	}

	// Convert string to byte slice and ensure null termination
	nameBytes := make([]byte, 16) // 15 chars + null terminator
	copy(nameBytes, name)

	// Call prctl syscall directly
	_, _, errno := syscall.RawSyscall(SYS_PRCTL, PR_SET_NAME, uintptr(unsafe.Pointer(&nameBytes[0])), 0)
	if errno != 0 {
		return errno
	}

	return nil
}

// setProcessNameArgv attempts to modify argv[0] to change the process title
func setProcessNameArgv(title string) error {
	// This is a best-effort attempt to modify argv[0]
	// On most Unix systems, we need to modify the actual memory location
	// where argv[0] is stored, not just the Go string slice

	if len(os.Args) == 0 {
		return fmt.Errorf("no command line arguments available")
	}

	// The approach here is to modify os.Args[0] in place
	// This works on some systems where the Go runtime preserves the original argv
	originalArgv0 := os.Args[0]
	maxLen := len(originalArgv0)

	// Truncate title if it's too long to fit in the original space
	newTitle := title
	if len(newTitle) >= maxLen {
		newTitle = title[:maxLen-1]
	}

	// Pad with spaces and null terminate to clear any remaining characters
	paddedTitle := newTitle
	for len(paddedTitle) < maxLen {
		paddedTitle += " "
	}

	// Try to update os.Args[0] - this may or may not affect ps output
	// depending on the system and Go runtime version
	os.Args[0] = paddedTitle

	// Note: For more reliable process title setting on Unix systems,
	// applications often use libraries like setproctitle() or manually
	// manipulate the argv memory region. However, this requires CGO
	// or unsafe memory operations that may not be portable.

	return nil
}
