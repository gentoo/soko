// Used to show the landing page of the application

package index

import (
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
)

// Show renders a template to show the landing page of the application
func Show(w http.ResponseWriter, r *http.Request) {
	count, _ := database.DBCon.Model((*models.Package)(nil)).Count()

	addedPackages := getAddedPackages(10)
	updatedVersions := getUpdatedVersions(10)

	renderIndexTemplate(w, createPageData(count, addedPackages, updatedVersions))
}
