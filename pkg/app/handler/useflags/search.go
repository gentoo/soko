// Used to search for USE flags

package useflags

import (
	"html/template"
	"net/http"
	utils2 "soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
)

// Search renders a template containing a list of search results
// for a given query of USE flags
func Search(w http.ResponseWriter, r *http.Request) {

	results, _ := r.URL.Query()["q"]

	param := results[0]

	var useflags []models.Useflag
	err := database.DBCon.Model(&useflags).Where("name LIKE ? ", (param + "%")).Select()
	if err != nil {
		panic(err)
	}

	data := struct {
		Page        string
		Search      string
		Useflags    []models.Useflag
		Application models.Application
	}{
		Page:        "useflags",
		Search:      param,
		Useflags:    useflags,
		Application: utils2.GetApplicationData(),
	}

	templates := template.Must(
		template.Must(
			template.New("Show").ParseGlob("web/templates/layout/*.tmpl")).
			ParseGlob("web/templates/useflags/search.tmpl"))

	templates.ExecuteTemplate(w, "search.tmpl", data)
}
