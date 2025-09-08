package anagram_test

import (
	"testing"

	anagram "github.com/phamnam2003/challenges/leetcode/242"
	"github.com/stretchr/testify/require"
)

func TestIsAnagram(t *testing.T) {
	t.Parallel()
	tCases := []struct {
		s        string
		t        string
		expected bool
	}{
		{
			s:        "anagram",
			t:        "nagaram",
			expected: true,
		},
		{
			s:        "rat",
			t:        "car",
			expected: false,
		},
	}

	for _, c := range tCases {
		require.Equal(t, c.expected, anagram.IsAnagram(c.s, c.t))
	}
}
