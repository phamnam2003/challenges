package fib

var hmap = make(map[int]int)

func Fib(n int) int {
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}

	if v, ok := hmap[n]; ok {
		return v
	} else {
		v := Fib(n-1) + Fib(n-2)
		hmap[n] = v
		return v
	}
}
