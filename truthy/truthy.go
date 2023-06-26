// Package truthy implements helpers to test the truthy-ness of the values.
package truthy

import "strings"

// nos is the list of well-known representations of `false`.
//
// Of course, if the user passes some weird string like `faooooolse` that would
// render it to be considered unjustifiably truthy. However we just can not deal
// with every possible corner case. So, let's just keep things simple.
//
// If need be after all, we can always extend this list (but maybe implement
// binary search if it grows too big).
var nos = [...]string{
	"",
	"0",
	"f",
	"false",
	"n",
	"no",
}

// Is returns `false` if the argument sounds like "false" (empty string, "0",
// "f", "false", and so on), and `true` otherwise.
func Is(val string) bool {
	val = strings.ToLower(val)

	for _, no := range nos {
		if val == no {
			return false
		}
	}

	return true
}
