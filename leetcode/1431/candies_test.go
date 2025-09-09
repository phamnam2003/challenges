package candies_test

import (
	"testing"

	candies "github.com/phamnam2003/challenges/leetcode/1431"
	"github.com/stretchr/testify/require"
)

func TestKidsWithCandies(t *testing.T) {
	tCases := []struct {
		candies      []int
		extraCandies int
		output       []bool
	}{
		{
			candies:      []int{2, 3, 5, 1, 3},
			extraCandies: 3,
			output:       []bool{true, true, true, false, true},
		},
		{
			candies:      []int{4, 2, 1, 1, 2},
			extraCandies: 1,
			output:       []bool{true, false, false, false, false},
		},
		{
			candies:      []int{12, 1, 12},
			extraCandies: 10,
			output:       []bool{true, false, true},
		},
	}

	for _, c := range tCases {
		t.Run("kids_with_candies", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.output, candies.KidsWithCandies(c.candies, c.extraCandies))
		})
	}
}
