package rob_test

import (
	"testing"

	rob "github.com/phamnam2003/challenges/leetcode/198"
	"github.com/stretchr/testify/require"
)

func TestRob(t *testing.T) {
	tCases := []struct {
		arr    []int
		expect int
	}{
		{
			arr:    []int{1, 2, 3, 1},
			expect: 4,
		},
		{
			arr:    []int{2, 7, 9, 3, 1},
			expect: 12,
		},
	}

	for _, c := range tCases {
		t.Run("robber", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, rob.Rob(c.arr), c.expect)
		})
	}
}
