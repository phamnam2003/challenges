package dest

func DestCity(paths [][]string) string {
	hmap := make(map[string]bool)

	for _, path := range paths {
		hmap[path[0]] = true
		hmap[path[1]] = hmap[path[1]]
	}

	for key, value := range hmap {
		if !value {
			return key
		}
	}
	return ""
}
