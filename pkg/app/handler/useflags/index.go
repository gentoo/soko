// Used to display the landing page of the USE flag section

package useflags

import (
	"html/template"
	"net/http"
	utils2 "soko/pkg/app/utils"
	"soko/pkg/models"
)

// Index renders a template to show the index page of the USE flags
// section containing a bubble chart of popular USE flags
func Index(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Header   models.Header
		Application models.Application
	}{
		Header:         models.Header{Title: "Useflags – ", Tab:   "useflags", },
		Application: utils2.GetApplicationData(),
	}

	templates := template.Must(
		template.Must(
			template.New("Show").ParseGlob("web/templates/layout/*.tmpl")).
			ParseGlob("web/templates/useflags/index.tmpl"))

	templates.ExecuteTemplate(w, "index.tmpl", data)
}
