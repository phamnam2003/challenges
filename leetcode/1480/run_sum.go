// Package runsum provides a function to calculate the running sum of an array.
package runsum

func RunningSum(nums []int) []int {
	for i := 1; i < len(nums); i++ {
		nums[i] += nums[i-1]
	}

	return nums
}
