// Used to show a specific package

package packages

import (
	"encoding/json"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/v9/orm"
	"html/template"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
	"time"
)

// Show renders a template to show a given package
func Show(w http.ResponseWriter, r *http.Request) {

	if strings.HasSuffix(r.URL.Path, "/changelog.html") {
		changelog(w, r)
		return
	} else if strings.HasSuffix(r.URL.Path, "/changelog.json") {
		changelogJSON(w, r)
		return
	} else if strings.HasSuffix(r.URL.Path, ".json") {
		buildJson(w, r)
		return
	}

	atom := r.URL.Path[len("/packages/"):]

	gpackage := new(models.Package)
	err := database.DBCon.Model(gpackage).
		Where("atom = ?", atom).
		Relation("Outdated").
		Relation("PkgCheckResults").
		Relation("Versions").
		Relation("Versions.Masks").
		Relation("Versions.PkgCheckResults").
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

// changelog renders a json version of the changelog
func changelogJSON(w http.ResponseWriter, r *http.Request) {

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

	var jsonChanges []Change
	for _, commit := range gpackage.Commits {

		var changedPackages []string
		var commitToPackages []models.CommitToPackage
		err = database.DBCon.Model(&commitToPackages).
			Where("commit_id = ?", commit.Id).
			Select()
		if err != nil {
			continue
		}
		for _, changedPackage := range commitToPackages {
			changedPackages = append(changedPackages, changedPackage.PackageAtom)
		}

		jsonChanges = append(jsonChanges, Change{
			Id:             commit.Id,
			AuthorName:     commit.AuthorName,
			AuthorEmail:    commit.AuthorEmail,
			AuthorDate:     commit.AuthorDate,
			CommitterName:  commit.CommitterName,
			CommitterEmail: commit.CommitterEmail,
			CommitterDate:  commit.CommitterDate,
			Message:        commit.Message,
			Files:          *commit.ChangedFiles,
			Packages:       changedPackages,
		})
	}

	jsonData := Changes{
		Changes: jsonChanges,
	}

	b, err := json.Marshal(jsonData)

	if err != nil {
		http.Error(w, "Internal Server Error",
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

type Changes struct {
	Changes []Change `json:"changes"`
}

type Change struct {
	Id             string              `json:"id"`
	AuthorName     string              `json:"author_name"`
	AuthorEmail    string              `json:"author_email"`
	AuthorDate     time.Time           `json:"author_date"`
	CommitterName  string              `json:"committer_name"`
	CommitterEmail string              `json:"committer_email"`
	CommitterDate  time.Time           `json:"committer_date"`
	Message        string              `json:"message"`
	Files          models.ChangedFiles `json:"files"`
	Packages       []string            `json:"package"`
}
