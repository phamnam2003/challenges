package kfreq

func TopKFrequent(nums []int, k int) []int {
	cmap := make(map[int]int)
	for _, v := range nums {
		cmap[v]++
	}

	buckets := make([][]int, len(nums)+1)
	for num, freq := range cmap {
		buckets[freq] = append(buckets[freq], num)
	}

	res := []int{}
	for i := len(buckets) - 1; i >= 0 && len(res) < k; i-- {
		if len(buckets[i]) > 0 {
			res = append(res, buckets[i]...)
		}
	}

	return res[:k]
}
