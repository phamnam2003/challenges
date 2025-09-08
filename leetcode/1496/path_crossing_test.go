package pcross_test

import (
	"testing"

	pcross "github.com/phamnam2003/challenges/leetcode/1496"
	"github.com/stretchr/testify/require"
)

func TestIsPathCrossing(t *testing.T) {
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
		require.Equal(t, pcross.IsPathCrossing(tc.path), tc.expt)
	}
}
