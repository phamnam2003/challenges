// Package pcross resolves the LeetCode problem 1496. Path Crossing
package pcross

import "fmt"

func IsPathCrossing(path string) bool {
	x, y := 0, 0
	vtNode := make(map[string]bool)

	// origin node visited
	vtNode["{0,0}"] = true

	for _, pos := range path {
		switch pos {
		case 'N':
			y++
		case 'S':
			y--
		case 'E':
			x++
		case 'W':
			x--
		}

		key := fmt.Sprintf("{%d,%d}", x, y)
		if _, ok := vtNode[key]; ok {
			return true
		} else {
			vtNode[key] = true
		}
	}
	return false
}
