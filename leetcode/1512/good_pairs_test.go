package gpairs_test

import (
	"testing"

	gpairs "github.com/phamnam2003/challenges/leetcode/1512"
	"github.com/stretchr/testify/require"
)

func TestNumIdenticalPairs(t *testing.T) {
	tCases := []struct {
		arrs     []int
		expected int
	}{
		{
			arrs:     []int{1, 2, 3, 1, 1, 3},
			expected: 4,
		},
		{
			arrs:     []int{1, 1, 1, 1},
			expected: 6,
		},
		{
			arrs:     []int{1, 2, 3},
			expected: 0,
		},
		{
			arrs:     []int{1, 2, 1, 2, 1, 2},
			expected: 6,
		},
	}

	for _, c := range tCases {
		count := gpairs.NumIdenticalPairs(c.arrs)
		require.Equal(t, c.expected, count)
	}
}
