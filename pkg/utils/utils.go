// SPDX-License-Identifier: GPL-2.0-only
package utils

import (
	"slices"
	"strings"
)

// Deduplicate accepts a slice of strings and returns
// a slice which only contains unique items.
func Deduplicate(items []string) []string {
	if len(items) > 1 {
		slices.Sort(items)
		return slices.Compact(items)
	} else {
		return items
	}
}

func SliceTrimSpaces(items []string) (res []string) {
	res = make([]string, len(items))
	for i, item := range items {
		res[i] = strings.TrimSpace(item)
	}
	return res
}
