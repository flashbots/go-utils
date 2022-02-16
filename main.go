package goutils

import (
	"os"
	"strconv"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func GetEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func GetEnvInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		val, err := strconv.Atoi(value)
		if err == nil {
			return val
		}
	}
	return defaultValue
}
