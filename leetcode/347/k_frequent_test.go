package kfreq_test

import (
	"testing"

	kfreq "github.com/phamnam2003/challenges/leetcode/347"
	"github.com/stretchr/testify/require"
)

func TestTopKFrequent(t *testing.T) {
	t.Parallel()
	tCases := []struct {
		array []int
		k     int
		res   []int
	}{}
	for _, c := range tCases {
		t.Run("k_freq", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, kfreq.TopKFrequent(c.array, c.k), c.res)
		})
	}
}
