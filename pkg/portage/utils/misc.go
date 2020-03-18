// Contains miscellaneous utility functions

package utils

// Contains tells whether string a contains string x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
