// Package strstack implements a simple stack for strings.
package strstack

func BuildArray(target []int, n int) []string {
	steps := make([]string, 0)
	mediate := 0

	for _, val := range target {
		for mediate < val-1 {
			steps = append(steps, "Push", "Pop")
			mediate++
		}

		steps = append(steps, "Push")
		mediate++
	}

	return steps
}
