package array_test

import (
	"testing"

	array "github.com/phamnam2003/challenges/leetcode/53"
	"github.com/stretchr/testify/require"
)

func TestMaxSubArray(t *testing.T) {
	tCases := []struct {
		arr    []int
		output int
	}{
		{
			arr:    []int{-2, 1, -3, 4, -1, 2, 1, -5, 4},
			output: 6,
		},
		{
			arr:    []int{1},
			output: 1,
		},
		{
			arr:    []int{5, 4, -1, 7, 8},
			output: 23,
		},
	}

	for _, c := range tCases {
		require.Equal(t, array.MaxSubArray(c.arr), c.output)
	}
}
