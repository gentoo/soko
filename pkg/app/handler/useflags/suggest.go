// Used to create USE flag suggestions

package useflags

import (
	"encoding/json"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"

	"github.com/go-pg/pg/v10"
)

// Suggest returns a json encoded suggestions of
// USE flags based on the given query
func Suggest(w http.ResponseWriter, r *http.Request) {
	results, found := r.URL.Query()["q"]
	if !found || len(results[0]) < 1 {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}
	param := results[0]

	var suggestions struct {
		Results []struct {
			Id          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"results"`
	}

	err := database.DBCon.Model((*models.Useflag)(nil)).
		Column("id", "name", "description").
		Where("name LIKE ?", param+"%").
		Select(&suggestions.Results)
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	jsondata, err := json.Marshal(suggestions)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}
