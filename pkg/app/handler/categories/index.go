// SPDX-License-Identifier: GPL-2.0-only

// Used to show a all categories

package categories

import (
	"encoding/json"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
)

// build the json for the categories overview page
func JSONCategories(w http.ResponseWriter, r *http.Request) {

	var categories []*models.Category
	err := database.DBCon.Model(&categories).Order("name ASC").Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var jsonCategories []CategoryDescription
	for _, category := range categories {
		jsonCategories = append(jsonCategories, CategoryDescription{
			Name:        category.Name,
			Url:         "https://packages.gentoo.org/categories/" + category.Name + ".json",
			Description: category.Description,
		})
	}

	b, err := json.Marshal(jsonCategories)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)

}

type CategoryDescription struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
}
