package utils

import "os"

// GetEnv returns the value of the environment variable or a default value
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
