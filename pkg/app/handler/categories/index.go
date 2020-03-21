// Used to show a all categories

package categories

import (
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
)

// Index renders a template to show all categories
func Index(w http.ResponseWriter, r *http.Request) {

	var categories []*models.Category
	err := database.DBCon.Model(&categories).Order("name ASC").Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	renderCategoryTemplate("index", createCategoriesData(categories), w)
}
