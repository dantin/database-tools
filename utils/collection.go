package utils

func StringInSlice(s string, list []string) bool {
	for _, t := range list {
		if t == s {
			return true
		}
	}
	return false
}
