// Used to search for packages

package packages

import (
	"github.com/go-pg/pg"
	"html/template"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
)

// Search renders a template containing a list of search results
// for a given query of packages
func Search(w http.ResponseWriter, r *http.Request) {

	searchTerm := getParameterValue("q", r)

	var packages []models.Package
	err := database.DBCon.Model(&packages).
		Where("atom LIKE ? ", ("%" + searchTerm + "%")).
		Relation("Versions").
		Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	renderPackageTemplate("search",
		"search",
		template.FuncMap{},
		getSearchData(packages, searchTerm),
		w)
}