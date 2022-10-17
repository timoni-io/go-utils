package maps

func Compare[K, T comparable](a, b map[K]T) bool {
	if len(a) != len(b) {
		return true
	}

	for k, v := range a {
		if b[k] != v {
			return true
		}
	}

	return false
}
