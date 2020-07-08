// Used to show a specific package

package useflags

import (
	"github.com/go-pg/pg/v9"
	"html/template"
	"net/http"
	utils2 "soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	"soko/pkg/utils"
	"strings"
)

// Show renders a template to show a given USE flag
func Show(w http.ResponseWriter, r *http.Request) {

	useflagName := r.URL.Path[len("/useflags/"):]

	var useflags []models.Useflag
	err := database.DBCon.Model(&useflags).Where("name = ? ", useflagName).Select()
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

	var versions []models.Version
	err = database.DBCon.Model(&versions).Column("atom").Where("useflags::jsonb @> ?", "\""+useflagName+"\"").Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	var packages []string
	for _, version := range versions {
		packages = append(packages, version.Atom)
	}

	packages = utils.Deduplicate(packages)

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
		Application:   utils2.GetApplicationData(),
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

	var otheruseexpands []models.Useflag
	err := database.DBCon.Model(&otheruseexpands).Where("use_expand = ? ", useExpand.UseExpand).Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	var versions []models.Version
	err = database.DBCon.Model(&versions).Column("atom").Where("useflags::jsonb @> ?", "\""+useExpand.Name+"\"").Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	var packages []string
	for _, version := range versions {
		packages = append(packages, version.Atom)
	}

	packages = utils.Deduplicate(packages)

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
		OtherUseExpands: otheruseexpands,
		Packages:        packages,
		Application:     utils2.GetApplicationData(),
	}

	templates := template.Must(
		template.Must(
		template.Must(
			template.New("Show").Funcs(funcMap).ParseGlob("web/templates/layout/*.tmpl")).
			ParseGlob("web/templates/useflags/useflagsheader.tmpl")).
			ParseGlob("web/templates/useflags/showexpand.tmpl"))

	templates.ExecuteTemplate(w, "showexpand.tmpl", data)
}
