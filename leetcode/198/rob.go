// Package rob provides a solution to the House Robber problem.
package rob

// Rob calculates the maximum amount of money that can be robbed without alerting the police.
func Rob(nums []int) int {
	rob1, rob2 := 0, 0

	for _, val := range nums {
		rob1, rob2 = rob2, max(rob1+val, rob2)
	}
	return rob2
}
