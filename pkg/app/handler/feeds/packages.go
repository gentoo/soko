package feeds

import (
	"net/http"
	"text/template"
)

// Show renders a template to show a given package
func Packages(funcMap template.FuncMap, data interface{}, w http.ResponseWriter) {
	templates := template.Must(
		template.New("changes.atom").
			Funcs(funcMap).
			ParseGlob("web/templates/feeds/packages.atom.tmpl"))

	w.Header().Set("Content-Type", "application/atom+xml")
	templates.ExecuteTemplate(w, "packages.atom.tmpl", data)
}
