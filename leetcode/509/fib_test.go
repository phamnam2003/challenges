package fib_test

import (
	"testing"

	fib "github.com/phamnam2003/challenges/leetcode/509"
	"github.com/stretchr/testify/require"
)

func TestFib(t *testing.T) {
	tCases := []struct {
		pos      int
		expected int
	}{
		{
			pos:      2,
			expected: 1,
		},
		{
			pos:      3,
			expected: 2,
		},
		{
			pos:      4,
			expected: 3,
		},
		{
			pos:      5,
			expected: 5,
		},
		{
			pos:      6,
			expected: 8,
		},
		{
			pos:      7,
			expected: 13,
		},
	}

	for _, c := range tCases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, fib.Fib(c.pos), c.expected)
		})
	}
}
