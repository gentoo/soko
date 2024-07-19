// SPDX-License-Identifier: GPL-2.0-only
package categories

import (
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
)

func OutdatedFeed(w http.ResponseWriter, r *http.Request) {
	categoryName := r.PathValue("category")
	var outdated []models.OutdatedPackages
	err := database.DBCon.Model(&outdated).
		Where("atom LIKE ?", categoryName+"/%").
		Order("atom").
		Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	utils.OutdatedFeed(w, "https://packages.gentoo.org/categories/"+categoryName+"/outdated", "category "+categoryName, outdated)
}

func StabilizationFeed(w http.ResponseWriter, r *http.Request) {
	categoryName := r.PathValue("category")
	var results []*models.PkgCheckResult
	err := database.DBCon.Model(&results).
		Column("atom", "cpv", "message").
		Where("class = ?", "StableRequest").
		Where("atom LIKE ?", categoryName+"/%").
		OrderExpr("cpv").
		Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	utils.StabilizationFeed(w, "https://packages.gentoo.org/categories/"+categoryName+"/stabilization", "category "+categoryName, results)
}
