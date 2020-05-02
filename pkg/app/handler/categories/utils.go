// miscellaneous utility functions used for categories

package categories

import (
	"html/template"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/models"
	"strings"
)

// getCategoryName returns the name of the
// category based on the given URL
func getCategoryName(r *http.Request) string {
	return strings.ReplaceAll(r.URL.Path[len("/categories/"):], ".json", "")
}

// createCategoriesData creates the data used in
// the template to display all categories
func createCategoriesData(categories []*models.Category) interface{} {
	return struct {
		Header models.Header
		Categories  []*models.Category
		Application models.Application
	}{
		Header:         models.Header{Title: "Categories – ", Tab:   "packages", },
		Categories:  categories,
		Application: utils.GetApplicationData(),
	}
}

// createCategoriesData creates the data used in
// the template to display a specific category
func createCategoryData(category models.Category) interface{} {
	return struct {
		Header      models.Header
		Category    models.Category
		Application models.Application
	}{
		Header:         models.Header{Title: category.Name + " – ", Tab:   "packages", },
		Category:    category,
		Application: utils.GetApplicationData(),
	}
}

// renderIndexTemplate renders all templates used for the categories section
func renderCategoryTemplate(page string, data interface{}, w http.ResponseWriter) {
	templates := template.Must(
		template.Must(
			template.New(page).
				ParseGlob("web/templates/layout/*.tmpl")).
			ParseGlob("web/templates/categories/*.tmpl"))

	templates.ExecuteTemplate(w, page+".tmpl", data)
}
