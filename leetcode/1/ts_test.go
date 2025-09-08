package ts_test

import (
	"testing"

	ts "github.com/phamnam2003/challenges/leetcode/1"
	"github.com/stretchr/testify/require"
)

func TestTwoSum(t *testing.T) {
	t.Parallel()
	tCases := []struct {
		arr    []int
		target int
		result []int
	}{
		{
			arr:    []int{1, 2, 6, 7, 9},
			target: 9,
			result: []int{3, 1},
		},
		{
			arr:    []int{1, 2, 6, 7, 9},
			target: 12,
			result: []int{},
		},
	}

	for _, c := range tCases {
		require.Equal(t, ts.TwoSum(c.arr, c.target), c.result)
	}
}
