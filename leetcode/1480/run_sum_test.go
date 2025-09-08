package runsum_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunningSum(t *testing.T) {
	t.Parallel()
	tCases := []struct {
		in  []int
		out []int
	}{}

	for _, c := range tCases {
		t.Run("run_sum", func(t *testing.T) {
			require.Equal(t, c.in, c.out)
		})
	}
}
