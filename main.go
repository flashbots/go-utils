// Package goutils implements various reusable utility functions.
package goutils

import (
	"os"
	"strconv"
)

// CheckErr panics if err is not nil
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

// GetEnv returns the value of the environment variable named by key, or defaultValue if the environment variable doesn't exist
func GetEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// GetEnvInt returns the value of the environment variable named by key, or defaultValue if the environment variable
// doesn't exist or is not a valid integer
func GetEnvInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		val, err := strconv.Atoi(value)
		if err == nil {
			return val
		}
	}
	return defaultValue
}
