// miscellaneous utility functions used for the about pages

package about

import (
	"html/template"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/models"
)

// renderAboutTemplate renders a specific about template
func renderAboutTemplate(w http.ResponseWriter, r *http.Request, page string){
	templates :=	template.Must(
						template.Must(
							template.New(page).
								ParseGlob("web/templates/layout/*.tmpl")).
								ParseGlob("web/templates/about/" + page + ".tmpl"))

	templates.ExecuteTemplate(w, page + ".tmpl", getPageData())
}

// getPageData returns the data used
// in all about templates
func getPageData() interface{}{
	return struct {
		Page            string
		Application     models.Application
	}{
		Page:           "about",
		Application:    utils.GetApplicationData(),
	}
}
