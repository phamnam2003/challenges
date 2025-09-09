package strstack_test

import (
	"testing"

	strstack "github.com/phamnam2003/challenges/leetcode/1441"
	"github.com/stretchr/testify/require"
)

func TestBuildArray(t *testing.T) {
	tCases := []struct {
		arr      []int
		val      int
		expected []string
	}{
		{
			arr:      []int{1, 3},
			val:      3,
			expected: []string{"Push", "Push", "Pop", "Push"},
		},
		{
			arr:      []int{1, 2, 3},
			val:      3,
			expected: []string{"Push", "Push", "Push"},
		},
		{
			arr:      []int{1, 2},
			val:      4,
			expected: []string{"Push", "Push"},
		},
		{
			arr:      []int{2, 3, 4},
			val:      4,
			expected: []string{"Push", "Pop", "Push", "Push", "Push"},
		},
	}

	for _, c := range tCases {
		t.Run("arr_stack", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.expected, strstack.BuildArray(c.arr, c.val))
		})
	}
}
