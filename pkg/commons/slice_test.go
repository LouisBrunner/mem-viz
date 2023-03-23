package commons_test

import (
	"testing"

	"github.com/LouisBrunner/dsc-viz/pkg/commons"
	"github.com/stretchr/testify/assert"
)

func Test_MapSlice_string2int(t *testing.T) {
	in := []string{
		"one",
		"two",
		"three",
	}
	expected := []int{
		3,
		3,
		5,
	}
	out := commons.MapSlice(in, func(in string) int {
		return len(in)
	})
	assert.Equal(t, expected, out)
}
