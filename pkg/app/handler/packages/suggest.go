// Used to create package suggestions

package packages

import (
	"encoding/json"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
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
	if err != nil {
		panic(err)
	}


	type Result struct{
		Name    string `json:"name"`
		Category string `json:"category"`
		description string `json:"description"`
	}

	type Results struct{
		Results  []*Result `json:"results"`
	}

	var results []*Result

	for  _, gpackage := range packages{
		results = append(results, &Result{
			Name:        gpackage.Name,
			Category:    gpackage.Category,
			description: gpackage.Versions[0].Description,
		})
	}

	result := Results{
		Results:  results,
	}

	b, err := json.Marshal(result)

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
