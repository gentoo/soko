// Used to show a specific category

package categories

import (
	"encoding/json"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"

	"github.com/go-pg/pg/v10"
)

// Show renders a template to show a given category
func Show(w http.ResponseWriter, r *http.Request) {
	categoryName, pageUrl, found := strings.Cut(r.URL.Path[len("/categories/"):], "/")
	if !found {
		if strings.HasSuffix(r.URL.Path, ".json") {
			buildJson(w, r, strings.TrimSuffix(categoryName, ".json"))
			return
		}
	}

	category := new(models.Category)
	query := database.DBCon.Model(category).
		Where("category.name = ?", categoryName).
		Relation("PackagesInformation").
		Relation("Packages", func(q *pg.Query) (*pg.Query, error) {
			return q.Order("name ASC"), nil
		}).Relation("Packages.Versions")

	pageName := "packages"
	switch pageUrl {
	case "outdated":
		pageName = "outdated"
		query = query.Relation("Packages.Outdated")
	}

	err := query.Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	renderCategoryTemplate("show", createCategoryData(pageName, *category), w)
}

// build the json for the category
func buildJson(w http.ResponseWriter, r *http.Request, categoryName string) {

	category := new(models.Category)
	err := database.DBCon.Model(category).
		Where("name = ?", categoryName).
		Relation("Packages").
		Relation("Packages.Versions").
		Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	categoryPackages := getJSONPackages(category)

	jsonCategory := Category{
		Name:     category.Name,
		Href:     "https://packages.gentoo.org/categories/" + category.Name,
		Packages: categoryPackages,
	}

	b, err := json.Marshal(jsonCategory)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// get all maintainers of the package in a format
// that is intended to be used to convert it to json
func getJSONPackages(category *models.Category) []Package {
	var categoryPackages []Package
	for _, gpackage := range category.Packages {
		if len(gpackage.Versions) > 0 {
			var homepage string
			if len(gpackage.Versions[0].Homepage) > 0 {
				homepage = gpackage.Versions[0].Homepage[0]
			}
			categoryPackages = append(categoryPackages, Package{
				Name:        gpackage.Name,
				Href:        homepage,
				Description: gpackage.Versions[0].Description,
			})
		}
	}
	return categoryPackages
}

type Category struct {
	Name     string    `json:"name"`
	Href     string    `json:"href"`
	Packages []Package `json:"packages"`
}

type Package struct {
	Name        string `json:"name"`
	Href        string `json:"href"`
	Description string `json:"description"`
}
