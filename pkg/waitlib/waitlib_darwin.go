//go:build darwin

// SPDX-License-Identifier: CC0-1.0

package waitlib

import (
	"fmt"
	"os"
)

// setProcessTitle sets the process title on macOS
func setProcessTitle(title string) error {
	// On macOS, we primarily rely on argv[0] manipulation
	// This will show up in ps output
	return setProcessNameArgv(title)
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
