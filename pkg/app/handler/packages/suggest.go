// Used to create package suggestions

package packages

import (
	"encoding/json"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"

	"github.com/go-pg/pg"
)

// Suggest returns json encoded suggestions of
// packages based on the given query
func Suggest(w http.ResponseWriter, r *http.Request) {

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

	type Result struct {
		Name        string `json:"name"`
		Category    string `json:"category"`
		description string `json:"description"`
	}

	type Results struct {
		Results []*Result `json:"results"`
	}

	results := make([]*Result, len(packages))
	for i, gpackage := range packages {
		results[i] = &Result{
			Name:        gpackage.Name,
			Category:    gpackage.Category,
			description: gpackage.Versions[0].Description,
		}
	}

	result := Results{
		Results: results,
	}

	b, err := json.Marshal(result)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
