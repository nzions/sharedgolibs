package util

import "os"

// MustGetEnv returns the value of the environment variable named by key.
// If the variable is not set or empty, returns the fallback value.
// This function unifies the environment variable handling across allmytails and googleemu.
func MustGetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
