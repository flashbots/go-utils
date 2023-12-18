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
// from the environment variable. It returns error if the environment variable's
// value can not be resolved into definitive `true` or `false`.
func Bool(name string, defaultValue bool, usage string) (*bool, error) {
	var err error
	value := defaultValue
	env := flagToEnv(name)
	if raw := os.Getenv(env); raw != "" {
		if pValue, pErr := truthy.Is(raw); pErr == nil {
			value = pValue
		} else {
			err = fmt.Errorf("invalid boolean value \"%s\" for environment variable %s: %w", raw, env, pErr)
		}
	}
	return flag.Bool(name, value, usage+fmt.Sprintf(" (env \"%s\")", env)), err
}

// MustBool handles error (if any) returned by Bool according to the behaviour
// configured by `flag.CommandLine.ErrorHandling()` by either ignoring it,
// exiting the process with status code 2, or panicking.
func MustBool(name string, defaultValue bool, usage string) *bool {
	res, err := Bool(name, defaultValue, usage)
	if err != nil {
		switch flag.CommandLine.ErrorHandling() {
		case flag.ContinueOnError:
			// continue
		case flag.ExitOnError:
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		case flag.PanicOnError:
			panic(err)
		}
	}
	if res == nil { // should never happen, guard added for NilAway
		panic(fmt.Sprintf("MustBool res for '%s' is nil", name))
	}
	return res
}

// Int is a convenience wrapper for integer flag that picks its default value
// from the environment variable. It returns error if the environment variable's
// value can not be parsed into integer.
func Int(name string, defaultValue int, usage string) (*int, error) {
	var err error
	value := defaultValue
	env := flagToEnv(name)
	if raw := os.Getenv(env); raw != "" {
		if pValue, pErr := strconv.Atoi(raw); pErr == nil {
			value = pValue
		} else {
			err = fmt.Errorf("invalid integer value \"%s\" for environment variable %s: %w", raw, env, pErr)
		}
	}
	return flag.Int(name, value, usage+fmt.Sprintf(" (env \"%s\")", env)), err
}

// MustInt handles error (if any) returned by Int according to the behaviour
// configured by `flag.CommandLine.ErrorHandling()` by either ignoring it,
// exiting the process with status code 2, or panicking.
func MustInt(name string, defaultValue int, usage string) *int {
	res, err := Int(name, defaultValue, usage)
	if err != nil {
		switch flag.CommandLine.ErrorHandling() {
		case flag.ContinueOnError:
			// continue
		case flag.ExitOnError:
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		case flag.PanicOnError:
			panic(err)
		}
	}

	if res == nil { // should never happen, guard added for NilAway
		panic(fmt.Sprintf("MustInt res for '%s' is nil", name))
	}

	return res
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
