// SPDX-License-Identifier: GPL-2.0-only
package utils

import (
	"soko/pkg/models"
)

func CountBugsCategories(bugs []*models.Bug) (generalCount, stabilizationCount, keywordingCount int) {
	for _, bug := range bugs {
		switch bug.Component {
		case string(models.BugComponentVulnerabilities):
			continue
		case string(models.BugComponentStabilization):
			stabilizationCount++
		case string(models.BugComponentKeywording):
			keywordingCount++
		default:
			generalCount++
		}
	}
	return
}
