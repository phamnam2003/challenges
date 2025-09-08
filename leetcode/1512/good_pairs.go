// Package gpairs implements solution for LeetCode problem 1512: Number of Good Pairs.
package gpairs

func NumIdenticalPairs(nums []int) int {
	result := 0
	hashMap := make(map[int]int)

	for _, value := range nums {
		if count, ok := hashMap[value]; ok {
			result += count
			hashMap[value] += 1
		} else {
			hashMap[value] = 1
		}
	}

	return result
}
