// Used to show the landing page of the application

package index

import (
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
)

// Show renders a template to show the landing page of the application
func Show(w http.ResponseWriter, r *http.Request) {
	count, _ := database.DBCon.Model((*models.Package)(nil)).Count()

	var packagesList []models.Package
	if utils.GetUserPreferences(r).General.LandingPageLayout == "classic" {
		packagesList = getAddedPackages(10)
	} else {
		packagesList = getSearchHistoryPackages(r)
	}

	updatedVersions := getUpdatedVersions(10)

	renderIndexTemplate(w, createPageData(count, packagesList, updatedVersions, utils.GetUserPreferences(r)))
}
