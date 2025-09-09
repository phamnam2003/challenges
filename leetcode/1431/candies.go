package candies

func KidsWithCandies(candies []int, extraCandies int) []bool {
	maxCan := candies[0]
	result := make([]bool, len(candies))

	for _, can := range candies {
		maxCan = max(maxCan, can)
	}

	for index, can := range candies {
		result[index] = can+extraCandies >= maxCan
	}

	return result
}
