package useflags

import (
	"html/template"
	"net/http"
	utils2 "soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"

	"github.com/go-pg/pg"
)

func Expand(w http.ResponseWriter, r *http.Request) {
	var useflags []models.Useflag
	err := database.DBCon.Model(&useflags).
		Where("scope = 'use_expand'").
		Order("use_expand", "name").
		Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	data := struct {
		Header      models.Header
		Page        string
		Useflags    []models.Useflag
		Application models.Application
	}{
		Header:      models.Header{Title: "Use Expand" + " – ", Tab: "useflags"},
		Page:        "expand",
		Useflags:    useflags,
		Application: utils2.GetApplicationData(),
	}

	templates := template.Must(
		template.Must(
			template.Must(
				template.New("Show").ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/useflags/browseuseflagsheader.tmpl")).
			ParseGlob("web/templates/useflags/listexpand.tmpl"))

	templates.ExecuteTemplate(w, "listexpand.tmpl", data)
}
