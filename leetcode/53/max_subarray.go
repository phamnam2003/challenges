package array

func MaxSubArray(nums []int) int {
	maxSum := nums[0]
	currSum := 0

	for _, val := range nums {
		if currSum < 0 {
			currSum = 0
		}
		currSum += val
		maxSum = max(maxSum, currSum)
	}

	return maxSum
}
