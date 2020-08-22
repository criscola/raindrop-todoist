package utils

// SInt64Contains checks if a slice s of int64 contains k
func SInt64Contains(s []int64, k int64) bool {
	for _, e := range s {
		if k == e {
			return true
		}
	}
	return false
}

// SStringContains checks if a slice s of string contains k
func SStringContains(s []string, k string) bool {
	for _, e := range s {
		if k == e {
			return true
		}
	}
	return false
}

// RemoveFromSString removes a string r from a slice s
func RemoveFromSString(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}