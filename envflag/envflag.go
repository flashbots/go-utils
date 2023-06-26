// Package envflag is a wrapper for stdlib's flag that adds the environment
// variables as additional source of the values for flags.
package envflag

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/flashbots/go-utils/truthy"
)

// Bool is a convenience wrapper for boolean flag that picks its default value
// from the environment variable.
func Bool(name string, defaultValue bool, usage string) *bool {
	value := defaultValue
	env := flagToEnv(name)
	if raw := os.Getenv(env); raw != "" {
		value = truthy.Is(raw)
	}
	return flag.Bool(name, value, usage+fmt.Sprintf(" (env \"%s\")", env))
}

// Int is a convenience wrapper for integer flag that picks its default value
// from the environment variable.
func Int(name string, defaultValue int, usage string) *int {
	value := defaultValue
	env := flagToEnv(name)
	if raw := os.Getenv(env); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			value = parsed
		}
	}
	return flag.Int(name, value, usage+fmt.Sprintf(" (env \"%s\")", env))
}

// String is a convenience wrapper for string flag that picks its default value
// from the environment variable.
func String(name, defaultValue, usage string) *string {
	value := defaultValue
	env := flagToEnv(name)
	if raw := os.Getenv(env); raw != "" {
		value = raw
	}
	return flag.String(name, value, usage+fmt.Sprintf(" (env \"%s\")", env))
}

func flagToEnv(flag string) string {
	return strings.ToUpper(
		strings.ReplaceAll(flag, "-", "_"),
	)
}
