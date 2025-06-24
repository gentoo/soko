// SPDX-License-Identifier: GPL-2.0-only
// Used to show a specific package

package packages

import (
	"encoding/json"
	"fmt"
	"net/http"
	"soko/pkg/app/layout"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

// Show renders a template to show a given package
func Show(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")
	packageName := r.PathValue("package")
	pageName := r.PathValue("pageName")
	atom := category + "/" + packageName

	if pageName == "" && strings.HasSuffix(packageName, ".json") {
		buildJson(w, r)
		return
	}

	var currentSubTab string

	var gpackage models.Package
	query := database.DBCon.Model(&gpackage).
		Relation("Bugs").
		Relation("PullRequests").
		Relation("Versions", func(q *pg.Query) (*pg.Query, error) {
			// performs mostly correct ordering of versions, which is perfected by sortVersionsDesc
			return q.Order("version DESC"), nil
		}).
		Relation("Versions.Bugs")

	switch pageName {
	case "changelog":
		currentSubTab = "Changelog"
		query = query.Relation("Commits", func(q *pg.Query) (*pg.Query, error) {
			// here be dragons
			const template = (`'%[1]s', (SELECT ARRAY_AGG("%[1]s") ` +
				`FROM jsonb_array_elements(COALESCE(NULLIF(changed_files -> '%[1]s', 'null'), '[]')) AS "%[1]s" ` +
				`WHERE "%[1]s" ->> 'Path' LIKE ?)`)
			return q.Column("commit_to_package.*",
				"commit.id", "preceding_commits", "message",
				"author_name", "author_email", "author_date",
				"committer_name", "committer_email", "committer_date").
				ColumnExpr(("json_build_object(" +
					fmt.Sprintf(template, "Modified") + "," +
					fmt.Sprintf(template, "Added") + "," +
					fmt.Sprintf(template, "Deleted") +
					") AS changed_files"), atom+"/%", atom+"/%", atom+"/%").
				Order("preceding_commits DESC").
				Limit(50), nil
		})
	case "changelog.json":
		changelogJSON(w, r)
		return
	case "qa-report":
		currentSubTab = "QA report"
		query = query.
			Relation("PkgCheckResults", func(q *pg.Query) (*pg.Query, error) {
				return q.Where("version IS NULL").Order("class", "message"), nil
			}).
			Relation("Versions.PkgCheckResults", func(q *pg.Query) (*pg.Query, error) {
				return q.Order("version", "class", "message"), nil
			})
	case "pull-requests":
		atom = strings.ReplaceAll(atom, "/pull-requests", "")
		currentSubTab = "Pull requests"
	case "bugs":
		atom = strings.ReplaceAll(atom, "/bugs", "")
		currentSubTab = "Bugs"
	case "security":
		atom = strings.ReplaceAll(atom, "/security", "")
		currentSubTab = "Security"
	case "dependencies":
		atom = strings.ReplaceAll(atom, "/dependencies", "")
		currentSubTab = "Dependencies"
		query = query.Relation("Versions.Dependencies")
	case "reverse-dependencies":
		atom = strings.ReplaceAll(atom, "/reverse-dependencies", "")
		currentSubTab = "Reverse Dependencies"
		query = query.Relation("ReverseDependencies")
	case "", "overview":
		query = query.Relation("Outdated").
			Relation("Versions.Masks").
			Relation("Versions.Deprecates")
		currentSubTab = "Overview"
	default:
		http.NotFound(w, r)
		return
	}

	err := query.Where("atom = ?", atom).Select()
	if err != nil || len(gpackage.Versions) == 0 {
		var pkgmove models.PkgMove
		err = database.DBCon.Model(&pkgmove).Where("source = ?", atom).Select()
		if err == nil {
			http.Redirect(w, r, strings.Replace(r.URL.String(), atom, pkgmove.Destination, 1), http.StatusMovedPermanently)
			return
		}

		http.NotFound(w, r)
		return
	}

	sortVersionsDesc(gpackage.Versions)

	layout.Layout(gpackage.Atom, layout.Packages, show(&gpackage, currentSubTab)).Render(r.Context(), w)
}

// changelog renders a json version of the changelog
func changelogJSON(w http.ResponseWriter, r *http.Request) {
	atom := getAtom(r)
	gpackage := new(models.Package)
	err := database.DBCon.Model(gpackage).
		Where("atom = ?", atom).
		Relation("Commits", func(q *pg.Query) (*pg.Query, error) {
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

func countBugs(gpackage *models.Package) (securityBugs, nonSecurityBugs int) {
	for _, bug := range gpackage.Bugs {
		if bug.Component == string(models.BugComponentVulnerabilities) {
			securityBugs++
		} else {
			nonSecurityBugs++
		}
	}
	return
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
