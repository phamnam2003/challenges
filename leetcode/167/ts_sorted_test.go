package tssorted_test

import (
	"testing"

	tssorted "github.com/phamnam2003/challenges/leetcode/167"
	"github.com/stretchr/testify/require"
)

func TestTwoSumSorted(t *testing.T) {
	t.Parallel()
	tCases := []struct {
		arr    []int
		target int
		result []int
	}{
		{
			arr:    []int{2, 7, 11, 15},
			target: 9,
			result: []int{1, 2},
		},
		{
			arr:    []int{2, 3, 4},
			target: 6,
			result: []int{1, 3},
		},
	}

	for _, c := range tCases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tssorted.TwoSum(c.arr, c.target), c.result)
		})
	}
}
