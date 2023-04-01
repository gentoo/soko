// Used to search for USE flags

package useflags

import (
	"html/template"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"

	"github.com/go-pg/pg"
)

// Search renders a template containing a list of search results
// for a given query of USE flags
func Search(w http.ResponseWriter, r *http.Request) {

	results := r.URL.Query()["q"]

	param := ""
	var useflags []models.Useflag
	if len(results) != 0 {
		param = results[0]
		err := database.DBCon.Model(&useflags).
			Column("name", "description", "scope", "package").
			Where("name LIKE ?", param+"%").
			OrderExpr("scope, name <-> ?", param).
			Select()
		if err != nil && err != pg.ErrNoRows {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
	}

	data := struct {
		Header      models.Header
		Page        string
		Search      string
		Useflags    []models.Useflag
		Application models.Application
	}{
		Header:      models.Header{Title: param + " – ", Tab: "useflags"},
		Page:        "search",
		Search:      param,
		Useflags:    useflags,
		Application: utils.GetApplicationData(),
	}

	templates := template.Must(
		template.Must(
			template.Must(
				template.New("Show").ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/useflags/browseuseflagsheader.tmpl")).
			ParseGlob("web/templates/useflags/search.tmpl"))

	templates.ExecuteTemplate(w, "search.tmpl", data)
}
