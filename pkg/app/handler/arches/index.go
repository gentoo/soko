package arches

import (
	"html/template"
	"net/http"
)

// Index renders a template to show a the landing page containing links to all arches feeds
func Index(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(
		template.Must(
			template.Must(
				template.New("index").
					ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/arches/archesheader.tmpl")).
			ParseGlob("web/templates/arches/index.tmpl"))

	templates.ExecuteTemplate(w, "index.tmpl", getPageData())
}
