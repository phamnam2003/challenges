package ts

func TwoSum(nums []int, target int) []int {
	hmap := make(map[int]int)

	for i, val := range nums {
		comple := target - val
		if j, ok := hmap[comple]; ok {
			return []int{i, j}
		} else {
			hmap[val] = i
		}
	}
	return []int{}
}
