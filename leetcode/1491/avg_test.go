package avg_test

import (
	"testing"

	avg "github.com/phamnam2003/challenges/leetcode/1491"
	"github.com/stretchr/testify/require"
)

func TestAverage(t *testing.T) {
	t.Parallel()
	tCases := []struct {
		salary []int
		output float64
	}{
		{
			salary: []int{4000, 3000, 1000, 2000},
			output: 2500.00000,
		},
		{
			salary: []int{1000, 2000, 3000},
			output: 2000.00000,
		},
		{
			salary: []int{6000, 5000, 4000, 3000, 2000, 1000},
			output: 3500.00000,
		},
		{
			salary: []int{48000, 59000, 99000, 13000, 78000, 45000, 31000, 17000, 39000, 37000, 93000, 77000, 33000, 28000, 4000, 54000, 67000, 6000, 1000, 11000},
			output: 41111.11111111111,
		},
	}

	for _, c := range tCases {
		t.Run("salary", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.output, avg.Average(c.salary))
		})
	}
}
