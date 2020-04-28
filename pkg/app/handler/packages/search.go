// Used to search for packages

package packages

import (
	"github.com/go-pg/pg"
	"html/template"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
)

// Search renders a template containing a list of search results
// for a given query of packages
func Search(w http.ResponseWriter, r *http.Request) {

	searchTerm := getParameterValue("q", r)
	searchQuery := buildSearchQuery(searchTerm)

	var packages []models.Package
	err := database.DBCon.Model(&packages).
		Where(searchQuery).
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

func buildSearchQuery(searchString string) string {
	var searchClauses []string
	for _, searchTerm := range strings.Split(searchString, " "){
		searchClauses = append(searchClauses,
			"(ARRAY[atom, category, name] && ARRAY['" + searchTerm +"'] OR (maintainers @> '[{\"Name\": \"" + searchTerm + "\"}]' OR maintainers @> '[{\"Email\": \"" + searchTerm + "\"}]'))")
	}
	return strings.Join(searchClauses, " AND ")
}
