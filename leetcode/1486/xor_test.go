package xor_test

import (
	"testing"

	xor "github.com/phamnam2003/challenges/leetcode/1486"
	"github.com/stretchr/testify/require"
)

func TestXOR(t *testing.T) {
	t.Parallel()
	tCases := []struct {
		n     int
		start int
		out   int
	}{
		{
			n:     5,
			start: 0,
			out:   8,
		},
		{
			n:     4,
			start: 3,
			out:   8,
		},
	}

	for _, c := range tCases {
		t.Run("xor", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.out, xor.XOR(c.n, c.start))
		})
	}
}
