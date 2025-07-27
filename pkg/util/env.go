// SPDX-License-Identifier: CC0-1.0

package util

import "os"

// Version is the current version of the util package
const Version = "0.1.0"

// MustGetEnv returns the value of the environment variable named by key.
// If the variable is not set or empty, returns the fallback value.
// Unifies environment variable handling across projects.
// Example usage:
//
//	dbURL := util.MustGetEnv("DATABASE_URL", "localhost:5432")
func MustGetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
