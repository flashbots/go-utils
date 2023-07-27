// Package truthy implements helpers to test the truthy-ness of the values.
package truthy

import (
	"fmt"
	"strings"
)

var isTruthy = map[string]bool{
	// truthy
	"1":    true,
	"t":    true,
	"true": true,
	"y":    true,
	"yes":  true,
	// non-truthy
	"":      false,
	"0":     false,
	"f":     false,
	"false": false,
	"n":     false,
	"no":    false,
}

// Is returns `false` if the argument sounds like "false" (empty string, "0",
// "f", "false", and so on), and `true` otherwise.
func Is(val string) (bool, error) {
	if res, known := isTruthy[strings.ToLower(val)]; known {
		return res, nil
	}
	return false, fmt.Errorf("can not resolve truthy-ness of \"%s\"", val)
}

// TrueOnError returns true if err is not nil, otherwise it returns res.
func TrueOnError(res bool, err error) bool {
	if err != nil {
		return true
	}
	return res
}

// FalseOnError returns false if err is not nil, otherwise it returns res.
func FalseOnError(res bool, err error) bool {
	if err != nil {
		return false
	}
	return res
}
