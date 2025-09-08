// Package xor provides a function to compute the XOR of a sequence of numbers.
package xor

func XOR(n, start int) int {
	result := start
	for i := 1; i < n; i++ {
		result ^= start + 2*i
	}

	return result
}
