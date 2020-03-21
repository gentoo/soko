// Used to create USE flag suggestions

package useflags

import (
	"encoding/json"
	"github.com/go-pg/pg/v9"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
)

// Suggest returns a json encoded suggestions of
// USE flags based on the given query
func Suggest(w http.ResponseWriter, r *http.Request) {

	results, _ := r.URL.Query()["q"]

	param := results[0]

	var useflags []models.Useflag
	err := database.DBCon.Model(&useflags).Where("name LIKE ? ", (param + "%")).Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	type Suggest struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	type Suggestions struct {
		Results []Suggest `json:"results"`
	}

	var suggests []Suggest

	for _, useflag := range useflags {
		suggests = append(suggests, Suggest{
			Id:          useflag.Id,
			Name:        useflag.Name,
			Description: useflag.Description,
		})
	}

	suggestions := Suggestions{
		Results: suggests,
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
