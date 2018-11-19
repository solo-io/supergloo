package util

func Contains(a []string, s string) bool {
	for _, n := range a {
		if s == n {
			return true
		}
	}
	return false
}
