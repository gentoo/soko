// Used to show a specific package

package packages

import (
	b64 "encoding/base64"
	"encoding/json"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

// Show renders a template to show a given package
func Show(w http.ResponseWriter, r *http.Request) {

	if strings.HasSuffix(r.URL.Path, "/changelog.json") {
		changelogJSON(w, r)
		return
	} else if strings.HasSuffix(r.URL.Path, ".json") {
		buildJson(w, r)
		return
	}

	atom := r.URL.Path[len("/packages/"):]
	pageName := "overview"
	userPreferences := utils.GetUserPreferences(r)

	if userPreferences.General.LandingPageLayout == "full" {
		updateSearchHistory(atom, w, r)
	}

	if strings.HasSuffix(r.URL.Path, "/changelog") {
		atom = strings.ReplaceAll(atom, "/changelog", "")
		pageName = "changelog"
	} else if strings.HasSuffix(r.URL.Path, "/qa-report") {
		atom = strings.ReplaceAll(atom, "/qa-report", "")
		pageName = "qa-report"
	} else if strings.HasSuffix(r.URL.Path, "/pull-requests") {
		atom = strings.ReplaceAll(atom, "/pull-requests", "")
		pageName = "pull-requests"
	} else if strings.HasSuffix(r.URL.Path, "/bugs") {
		atom = strings.ReplaceAll(atom, "/bugs", "")
		pageName = "bugs"
	} else if strings.HasSuffix(r.URL.Path, "/security") {
		atom = strings.ReplaceAll(atom, "/security", "")
		pageName = "security"
	} else if strings.HasSuffix(r.URL.Path, "/dependencies") {
		atom = strings.ReplaceAll(atom, "/dependencies", "")
		pageName = "dependencies"
	} else if strings.HasSuffix(r.URL.Path, "/reverse-dependencies") {
		atom = strings.ReplaceAll(atom, "/reverse-dependencies", "")
		pageName = "reverse-dependencies"
	}

	var gpackage models.Package
	err := database.DBCon.Model(&gpackage).
		Where("atom = ?", atom).
		Relation("Outdated").
		Relation("PkgCheckResults").
		Relation("Bugs").
		Relation("PullRequests").
		Relation("Versions").
		Relation("Versions.Bugs").
		Relation("Versions.Masks").
		Relation("Versions.Deprecates").
		Relation("Versions.PkgCheckResults").
		Relation("Versions.Dependencies").
		Relation("ReverseDependencies").
		Relation("Commits", func(q *pg.Query) (*pg.Query, error) {
			return q.Order("preceding_commits DESC").Limit(userPreferences.Packages.Overview.ChangelogLength), nil
		}).
		Select()

	if err != nil {
		http.NotFound(w, r)
		return
	}

	sortVersionsDesc(gpackage.Versions)

	localUseflags, globalUseflags, useExpands := getPackageUseflags(&gpackage)
	securityBugs, nonSecurityBugs := countBugs(&gpackage)

	renderPackageTemplate("show",
		"*",
		GetFuncMap(),
		createPackageData(pageName, &gpackage, localUseflags, globalUseflags, useExpands, userPreferences, securityBugs, nonSecurityBugs),
		w)
}

func updateSearchHistory(atom string, w http.ResponseWriter, r *http.Request) {
	var cookie, err = r.Cookie("search_history")
	var packages string
	if err == nil {
		cookieValue, err := b64.StdEncoding.DecodeString(cookie.Value)
		if err == nil {
			packagesList := strings.Split(string(cookieValue), ",")
			if strings.Contains(string(cookieValue), atom) {
				newPackagesList := []string{}
				for _, gpackage := range packagesList {
					if gpackage != atom {
						newPackagesList = append(newPackagesList, gpackage)
					}
				}
				packagesList = newPackagesList
			}
			packagesList = append(packagesList, atom)
			if len(packagesList) > 10 {
				packagesList = packagesList[len(packagesList)-10:]
			}
			packages = strings.Join(packagesList, ",")
		} else {
			packages = atom
		}
	} else {
		packages = atom
	}

	updatedCookie := http.Cookie{
		Name:    "search_history",
		Path:    "/",
		Value:   b64.StdEncoding.EncodeToString([]byte(packages)),
		Expires: time.Now().Add(365 * 24 * time.Hour),
	}
	http.SetCookie(w, &updatedCookie)
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
		if bug.Component == "Vulnerabilities" {
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
