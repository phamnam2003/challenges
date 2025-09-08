// Package avg provides a function to calculate the average salary excluding the minimum and maximum salaries.
package avg

func Average(salary []int) float64 {
	min := salary[0]
	max := min
	sum := 0

	for _, sal := range salary {
		if sal > max {
			max = sal
		} else if sal < min {
			min = sal
		}
		sum += sal
	}
	sum = sum - min - max
	return float64(sum) / float64(len(salary)-2)
}
