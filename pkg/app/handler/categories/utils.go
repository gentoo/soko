// miscellaneous utility functions used for categories

package categories

import (
	"html/template"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/models"
)

// createCategoriesData creates the data used in
// the template to display all categories
func createCategoriesData(categories []*models.Category) interface{} {
	return struct {
		Header      models.Header
		Name        string
		Categories  []*models.Category
		Application models.Application
	}{
		Header:      models.Header{Title: "Categories – ", Tab: "packages"},
		Name:        "Categories",
		Categories:  categories,
		Application: utils.GetApplicationData(),
	}
}

// createCategoriesData creates the data used in
// the template to display a specific category
func createCategoryData(pageName string, category models.Category, pullRequests []models.GithubPullRequest) interface{} {
	return struct {
		PageName     string
		Header       models.Header
		Category     models.Category
		PullRequests []models.GithubPullRequest
		Application  models.Application
	}{
		PageName:     pageName,
		Header:       models.Header{Title: category.Name + " – ", Tab: "packages"},
		Category:     category,
		PullRequests: pullRequests,
		Application:  utils.GetApplicationData(),
	}
}

// renderIndexTemplate renders all templates used for the categories section
func renderCategoryTemplate(page string, data interface{}, w http.ResponseWriter) {
	templates := template.Must(
		template.Must(
			template.Must(
				template.New(page).
					Funcs(template.FuncMap{
						"add": func(a, b int) int {
							return a + b
						},
					}).
					ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/categories/components/*.tmpl")).
			ParseGlob("web/templates/categories/*.tmpl"))

	templates.ExecuteTemplate(w, page+".tmpl", data)
}
