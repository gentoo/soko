package feeds

import (
	"text/template"
	"net/http"
	"soko/pkg/logger"
)

// Show renders a template to show a given package
func Changes(funcMap template.FuncMap, data interface{}, w http.ResponseWriter) {
	logger.Info.Println("Changes atom template")
	templates := template.Must(
				template.New("changes.atom").
					Funcs(funcMap).
					ParseGlob("web/templates/feeds/changes.atom.tmpl"))

	//w.Header().Set("Content-Type", "application/atom+xml")
	templates.ExecuteTemplate(w, "changes.atom.tmpl", data)
}

