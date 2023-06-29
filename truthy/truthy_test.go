package truthy_test

import (
	"fmt"
	"testing"

	"github.com/flashbots/go-utils/truthy"
	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	{ // truthy values
		for _, y := range []string{
			"1",
			"t",
			"T",
			"true",
			"True",
			"Y",
			"yes",
		} {
			assert.True(
				t,
				truthy.FalseOnError(truthy.Is(y)),
				fmt.Sprintf("Value '%s' must render as truthy", y),
			)
		}
	}
	{ // falsy values
		for _, n := range []string{
			"",
			"0",
			"f",
			"F",
			"false",
			"False",
			"N",
			"no",
		} {
			assert.False(
				t,
				truthy.TrueOnError(truthy.Is(n)),
				fmt.Sprintf("Value '%s' must render as falsy", n),
			)
		}
	}
}
