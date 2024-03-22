package utils

import (
	"soko/pkg/models"
)

func CountBugsCategories(bugs []*models.Bug) (generalCount, stabilizationCount, keywordingCount int) {
	for _, bug := range bugs {
		switch bug.Component {
		case "Current packages":
			generalCount++
		case "Stabilization":
			stabilizationCount++
		case "Keywording":
			keywordingCount++
		default:
			continue
		}
	}
	return
}
