//go:build !linux && !darwin

// SPDX-License-Identifier: CC0-1.0

package waitlib

// setProcessTitle is a no-op on unsupported platforms
func setProcessTitle(title string) error {
	// For other platforms, just silently succeed
	// Process title setting is not supported
	return nil
}
