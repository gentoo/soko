// Used to show a specific category

package categories

import (
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
)

// Show renders a template to show a given category
func Show(w http.ResponseWriter, r *http.Request) {

	category := new(models.Category)
	err := database.DBCon.Model(category).
		Where("name = ?", getCategoryName(r)).
		Relation("Packages").
		Relation("Packages.Versions").
		Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	renderCategoryTemplate("show", createCategoryData(*category), w)
}
