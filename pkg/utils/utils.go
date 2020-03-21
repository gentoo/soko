package utils

import "sort"

// Deduplicate accepts a slice of strings and returns
// a slice which only contains unique items.
func Deduplicate(items []string) []string {
	if items != nil && len(items) > 1 {
		sort.Strings(items)
		j := 0
		for i := 1; i < len(items); i++ {
			if items[j] == items[i] {
				continue
			}
			j++
			// preserve the original data
			// in[i], in[j] = in[j], in[i]
			// only set what is required
			items[j] = items[i]
		}
		result := items[:j+1]
		return result
	} else {
		return items
	}
}
