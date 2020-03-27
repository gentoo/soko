// Used to show a specific package

package packages

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/v9/orm"
	"html/template"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
)

// Show renders a template to show a given package
func Show(w http.ResponseWriter, r *http.Request) {

	if strings.HasSuffix(r.URL.Path, "/changelog.html") {
		changelog(w, r)
		return
	}

	atom := r.URL.Path[len("/packages/"):]

	gpackage := new(models.Package)
	err := database.DBCon.Model(gpackage).
		Where("atom = ?", atom).
		Relation("Versions").
		Relation("Versions.Masks").
		Select()

	if err != nil {
		http.NotFound(w, r)
		return
	}

	sortVersionsDesc(gpackage.Versions)

	localUseflags, globalUseflags, useExpands := getPackageUseflags(gpackage)

	renderPackageTemplate("show",
		"*",
		GetFuncMap(),
		createPackageData(gpackage, localUseflags, globalUseflags, useExpands),
		w)
}

// changelog renders a template to show the changelog of a given package
func changelog(w http.ResponseWriter, r *http.Request) {

	atom := getAtom(r)
	gpackage := new(models.Package)
	err := database.DBCon.Model(gpackage).
		Where("atom = ?", atom).
		Relation("Commits", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("preceding_commits DESC").Limit(5), nil
		}).
		Select()

	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	templates := template.Must(
		template.New("Changelog").
			Funcs(GetFuncMap()).
			ParseGlob("web/templates/packages/changelog/*.tmpl"))

	templates.ExecuteTemplate(w, "changelog.tmpl", getChangelogData(gpackage.Commits, atom))
}
