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

	var suggestions struct {
		Results []struct {
			Name        string `json:"name"`
			Category    string `json:"category"`
			Description string `json:"description"`
		} `json:"results"`
	}

	descriptionQuery := database.DBCon.Model((*models.Version)(nil)).
		Column("description").
		Where("atom = package.atom").
		Limit(1)
	err := database.DBCon.Model((*models.Package)(nil)).
		Column("name", "category").
		ColumnExpr("(?) AS description", descriptionQuery).
		Where("atom LIKE ? ", "%"+searchTerm+"%").
		OrderExpr("name <-> ?", searchTerm).
		Limit(10).
		Select(&suggestions.Results)
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(suggestions)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
