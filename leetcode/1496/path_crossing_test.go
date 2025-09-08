package pcross_test

import (
	"testing"

	pcross "github.com/phamnam2003/challenges/leetcode/1496"
	"github.com/stretchr/testify/require"
)

func TestIsPathCrossing(t *testing.T) {
	t.Parallel()
	tCases := []struct {
		path string
		expt bool
	}{
		{
			path: "NES",
			expt: false,
		},
		{
			path: "NESWW",
			expt: true,
		},
		{
			path: "NNSWWEWSSESSWENNW",
			expt: true,
		},
		{
			path: "ENNNNNNNNNNNEEEEEEEEEESSSSSSSSSS",
			expt: false,
		},
	}

	for _, tc := range tCases {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, pcross.IsPathCrossing(tc.path), tc.expt)
		})
	}
}
