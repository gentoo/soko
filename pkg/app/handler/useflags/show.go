// Used to show a specific package

package useflags

import (
	"html/template"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"

	"github.com/go-pg/pg/v10"
)

// Show renders a template to show a given USE flag
func Show(w http.ResponseWriter, r *http.Request) {
	useFlagName := r.URL.Path[len("/useflags/"):]

	var useflags []models.Useflag
	err := database.DBCon.Model(&useflags).Where("name = ?", useFlagName).Select()
	if err != nil || len(useflags) < 1 {
		http.NotFound(w, r)
		return
	}

	useflag := useflags[0]
	var localuseflags []models.Useflag

	for _, use := range useflags {
		if use.Scope == "global" {
			useflag = use
		} else if use.Scope == "local" {
			localuseflags = append(localuseflags, use)
		} else if use.Scope == "use_expand" {
			ShowUseExpand(w, r, use)
			return
		}
	}

	var packages []string
	err = database.DBCon.Model((*models.Version)(nil)).
		Column("atom").Distinct().
		Where("useflags::jsonb @> ?", "\""+useFlagName+"\"").
		Select(&packages)
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	data := struct {
		Header        models.Header
		Page          string
		Useflag       models.Useflag
		LocalUseflags []models.Useflag
		Packages      []string
		Application   models.Application
	}{
		Header:        models.Header{Title: useflag.Name + " – ", Tab: "useflags"},
		Page:          "show",
		Useflag:       useflag,
		LocalUseflags: localuseflags,
		Packages:      packages,
		Application:   utils.GetApplicationData(),
	}

	templates := template.Must(
		template.Must(
			template.Must(
				template.New("Show").Funcs(template.FuncMap{
					"replaceall": strings.ReplaceAll,
				}).ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/useflags/useflagsheader.tmpl")).
			ParseGlob("web/templates/useflags/show.tmpl"))

	templates.ExecuteTemplate(w, "show.tmpl", data)
}

// ShowUseExpand renders a template to show a given use_expand
func ShowUseExpand(w http.ResponseWriter, r *http.Request, useExpand models.Useflag) {
	funcMap := template.FuncMap{
		"replaceall": strings.ReplaceAll,
	}

	var otherUseExpands []models.Useflag
	err := database.DBCon.Model(&otherUseExpands).
		Column("name", "description").
		Where("use_expand = ?", useExpand.UseExpand).
		Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	var packages []string
	err = database.DBCon.Model((*models.Version)(nil)).
		Column("atom").Distinct().
		Where("useflags::jsonb @> ?", "\""+useExpand.Name+"\"").
		Select(&packages)
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	data := struct {
		Header          models.Header
		Page            string
		Useflag         models.Useflag
		OtherUseExpands []models.Useflag
		Packages        []string
		Application     models.Application
	}{
		Header:          models.Header{Title: useExpand.Name + " – ", Tab: "useflags"},
		Page:            "show",
		Useflag:         useExpand,
		OtherUseExpands: otherUseExpands,
		Packages:        packages,
		Application:     utils.GetApplicationData(),
	}

	templates := template.Must(
		template.Must(
			template.Must(
				template.New("Show").Funcs(funcMap).ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/useflags/useflagsheader.tmpl")).
			ParseGlob("web/templates/useflags/showexpand.tmpl"))

	templates.ExecuteTemplate(w, "showexpand.tmpl", data)
}
