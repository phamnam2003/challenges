package dest_test

import (
	"testing"

	dest "github.com/phamnam2003/challenges/leetcode/1436"
	"github.com/stretchr/testify/require"
)

func TestDestCity(t *testing.T) {
	t.Parallel()
	tCases := []struct {
		path   [][]string
		output string
	}{
		{
			path:   [][]string{{"London", "New York"}, {"New York", "Lima"}, {"Lima", "Sao Paulo"}},
			output: "Sao Paulo",
		},
		{
			path:   [][]string{{"B", "C"}, {"D", "B"}, {"C", "A"}},
			output: "A",
		},
		{
			path:   [][]string{{"A", "Z"}},
			output: "Z",
		},
	}

	for _, c := range tCases {
		t.Run("dest_city", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.output, dest.DestCity(c.path))
		})
	}
}
