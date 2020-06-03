package graphiql

import (
	"html/template"
	"net/http"
	"soko/pkg/config"
)

func Show(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")

	templates := template.Must(
			template.New("graphiql").
				ParseGlob("web/templates/api/explore/*.tmpl"))

	templates.ExecuteTemplate(w, "graphiql.tmpl", template.URL(config.GraphiqlEndpoint()))
}
