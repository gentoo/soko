// SPDX-License-Identifier: GPL-2.0-only
package utils

import (
	"soko/pkg/utils"
	"strings"
)

// FormatRestricts returns a string containing a comma separated
// list of capitalized first letters of the package restricts
func FormatRestricts(restricts []string) string {
	var result []string
	for _, restrict := range restricts {
		if restrict != "" && restrict != "(" && restrict != ")" && !strings.HasSuffix(restrict, "?") {
			result = append(result, strings.ToUpper(string(restrict[0])))
		}
	}
	result = utils.Deduplicate(result)
	return strings.Join(result, ", ")
}
