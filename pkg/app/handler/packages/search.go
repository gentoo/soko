// SPDX-License-Identifier: GPL-2.0-only
// Used to search for packages

package packages

import (
	"encoding/json"
	"net/http"
	"soko/pkg/app/handler/feeds"
	"soko/pkg/app/layout"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"

	"github.com/go-pg/pg/v10"
)

type searchResults struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// Search renders a template containing a list of search results
// for a given query of packages
func Search(w http.ResponseWriter, r *http.Request) {
	searchTerm := getParameterValue("q", r)

	switch {
	case searchTerm == "":
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	case strings.Contains(searchTerm, "@"):
		var maintainers []models.Maintainer
		_ = database.DBCon.Model(&maintainers).Where("email = ?", searchTerm).Select()
		if len(maintainers) > 0 {
			http.Redirect(w, r, "/maintainer/"+searchTerm, http.StatusMovedPermanently)
			return
		}
	case len(searchTerm) > 0 && searchTerm[len(searchTerm)-1] == '/':
		categoryName := searchTerm[:len(searchTerm)-1]
		count, err := database.DBCon.Model((*models.Category)(nil)).Where("name = ?", categoryName).Count()
		if err == nil && count > 0 {
			http.Redirect(w, r, "/categories/"+categoryName, http.StatusMovedPermanently)
			return
		}
	case strings.Contains(searchTerm, "/"):
		var packages []models.Package
		_ = database.DBCon.Model(&packages).Where("atom = ?", searchTerm).Select()
		if len(packages) > 0 {
			http.Redirect(w, r, "/packages/"+searchTerm, http.StatusMovedPermanently)
			return
		}
	}

	var results []searchResults
	descriptionQuery := database.DBCon.Model((*models.Version)(nil)).
		Column("description").
		Where("atom = package.atom").
		Limit(1)
	query := database.DBCon.Model((*models.Package)(nil)).
		Column("name", "category").
		ColumnExpr("(?) AS description", descriptionQuery)

	if strings.Contains(searchTerm, "*") {
		// if the query contains wildcards
		wildcardSearchTerm := strings.ReplaceAll(searchTerm, "*", "%")
		query = query.
			WhereOr("atom LIKE ?", wildcardSearchTerm).
			WhereOr("name LIKE ?", wildcardSearchTerm)
	} else {
		// if the query contains no wildcards do a fuzzy search
		query = BuildSearchQuery(query, searchTerm).
			WhereOr("atom LIKE ?", "%"+searchTerm+"%")
	}

	err := query.OrderExpr("name <-> ?", searchTerm).
		Select(&results)
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}
	if len(results) == 1 {
		http.Redirect(w, r, "/packages/"+results[0].Category+"/"+results[0].Name, http.StatusMovedPermanently)
		return
	}

	layout.Layout(searchTerm, layout.Packages, search(searchTerm, results)).Render(r.Context(), w)
}

// Search renders a template containing a list of search results
// for a given query of packages
func SearchFeed(w http.ResponseWriter, r *http.Request) {

	searchTerm := getParameterValue("q", r)
	searchTerm = strings.ReplaceAll(searchTerm, "*", "")

	var packages []models.Package
	err := BuildSearchQuery(database.DBCon.Model(&packages), searchTerm).
		Relation("Versions").
		OrderExpr("name <-> ?", searchTerm).
		Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	feeds.Packages(searchTerm, packages, w)
}

func BuildSearchQuery(query *pg.Query, searchString string) *pg.Query {
	for _, searchTerm := range strings.Split(searchString, " ") {
		if searchTerm != "" {
			marshal, err := json.Marshal(searchTerm)
			if err != nil {
				continue
			}
			query = query.WhereGroup(func(q *pg.Query) (*pg.Query, error) {
				return q.WhereOr("category % ?", searchTerm).
					WhereOr("name % ?", searchTerm).
					WhereOr("atom % ?", searchTerm).
					WhereOr("maintainers @> ?", `[{"Name": `+string(marshal)+`}]`).
					WhereOr("maintainers @> ?", `[{"Email": `+string(marshal)+`}]`), nil
			})
		}
	}
	return query
}
