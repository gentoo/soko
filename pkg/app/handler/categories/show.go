// SPDX-License-Identifier: GPL-2.0-only
// Used to show a specific category

package categories

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-pg/pg/v10"

	"soko/pkg/app/handler/packages/components"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
)

func common(w http.ResponseWriter, r *http.Request) (categoryName string, category models.Category, err error) {
	categoryName = r.PathValue("category")

	err = database.DBCon.Model(&category).
		Where("category.name = ?", categoryName).
		Relation("PackagesInformation").Select()
	if err == pg.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	return
}

type packageInfo struct {
	Package     string
	Description string
}

func ShowPackages(w http.ResponseWriter, r *http.Request) {
	categoryName := r.PathValue("category")
	if strings.HasSuffix(categoryName, ".json") {
		buildJson(w, categoryName[:len(categoryName)-5])
		return
	}
	categoryName, category, err := common(w, r)
	if err != nil {
		return
	}

	var packages []packageInfo
	err = database.DBCon.Model((*models.Version)(nil)).
		DistinctOn("package").
		Column("package", "description").
		Where("category = ?", categoryName).
		Order("package ASC").
		Select(&packages)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	renderShowPage(w, r, "Packages", &category,
		showPackages(categoryName, packages))
}

func ShowOutdated(w http.ResponseWriter, r *http.Request) {
	categoryName, category, err := common(w, r)
	if err != nil {
		return
	}

	var outdated []components.OutdatedItem
	descriptionQuery := database.DBCon.Model((*models.Version)(nil)).
		Column("description").
		Where("atom = outdated_packages.atom").
		Limit(1)
	err = database.DBCon.Model((*models.OutdatedPackages)(nil)).
		Column("atom").ColumnExpr("(?) AS description", descriptionQuery).
		Where("atom LIKE ?", categoryName+"/%").
		Order("atom").
		Select(&outdated)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	renderShowPage(w, r, "Outdated", &category,
		components.Outdated(outdated))
}

func ShowPullRequests(w http.ResponseWriter, r *http.Request) {
	categoryName, category, err := common(w, r)
	if err != nil {
		return
	}

	var pullRequests []*models.GithubPullRequest
	err = database.DBCon.Model(&pullRequests).
		Join("JOIN package_to_github_pull_requests ON package_to_github_pull_requests.github_pull_request_id = github_pull_request.id").
		Where("package_to_github_pull_requests.package_atom LIKE ?", categoryName+"/%").
		Group("github_pull_request.id").
		Order("github_pull_request.created_at DESC").
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	renderShowPage(w, r, "Pull requests", &category,
		components.PullRequests(pullRequests))
}

func ShowBugs(w http.ResponseWriter, r *http.Request) {
	categoryName, category, err := common(w, r)
	if err != nil {
		return
	}
	var bugs []*models.Bug
	err = database.DBCon.Model(&bugs).
		DistinctOn("id::INT").
		Column("id", "summary", "component", "assignee").
		OrderExpr("id::INT").
		Where("id IN (?)",
			database.DBCon.Model((*models.PackageToBug)(nil)).
				Column("bug_id").
				Where("package_atom LIKE ?", categoryName+"/%")).
		WhereOr("id IN (?)",
			database.DBCon.Model((*models.VersionToBug)(nil)).
				Column("bug_id").
				Join("JOIN versions").JoinOn("version_id = versions.id").
				Where("versions.atom LIKE ?", categoryName+"/%")).
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	generalCount, stabilizationCount, keywordingCount := utils.CountBugsCategories(bugs)
	renderShowPage(w, r, "Bugs", &category,
		components.Bugs("", generalCount, stabilizationCount, keywordingCount, bugs))
}

func ShowSecurity(w http.ResponseWriter, r *http.Request) {
	categoryName, category, err := common(w, r)
	if err != nil {
		return
	}
	var bugs []*models.Bug
	err = database.DBCon.Model(&bugs).
		DistinctOn("id::INT").
		Column("id", "summary", "component", "assignee").
		OrderExpr("id::INT").
		Where("component = ?", models.BugComponentVulnerabilities).
		Where("id IN (?)",
			database.DBCon.Model((*models.PackageToBug)(nil)).
				Column("bug_id").
				Where("package_atom LIKE ?", categoryName+"/%")).
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	renderShowPage(w, r, "Security", &category,
		components.SecurityBugs("", bugs))
}

func ShowStabilizations(w http.ResponseWriter, r *http.Request) {
	categoryName, category, err := common(w, r)
	if err != nil {
		return
	}

	var results []*models.PkgCheckResult
	err = database.DBCon.Model(&results).
		Column("atom", "cpv", "message").
		Where("class = ?", "StableRequest").
		Where("atom LIKE ?", categoryName+"/%").
		OrderExpr("cpv").
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	renderShowPage(w, r, "Stabilization", &category,
		components.Stabilizations(results))
}

func ShowStabilizationFile(w http.ResponseWriter, r *http.Request) {
	categoryName := r.PathValue("category")
	var results []*models.PkgCheckResult
	err := database.DBCon.Model(&results).
		Column("category", "package", "version", "message").
		Where("class = ?", "StableRequest").
		Where("atom LIKE ?", categoryName+"/%").
		OrderExpr("cpv").
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	pageName := r.URL.Path[strings.LastIndexByte(r.URL.Path, '/')+1:]
	utils.StabilizationExport(w, pageName, results)
}

// build the json for the category
func buildJson(w http.ResponseWriter, categoryName string) {
	var jsonCategory struct {
		Name     string `json:"name"`
		Href     string `json:"href"`
		Packages []struct {
			Package     string `json:"name"`
			Href        string `json:"href"`
			Description string `json:"description"`
		} `json:"packages"`
	}

	err := database.DBCon.Model((*models.Version)(nil)).
		DistinctOn("package").
		Column("package", "description").
		ColumnExpr("homepage ->> 0 AS href").
		Where("category = ?", categoryName).
		Order("package ASC").
		Select(&jsonCategory.Packages)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonCategory.Name = categoryName
	jsonCategory.Href = "https://packages.gentoo.org/categories/" + categoryName

	b, err := json.Marshal(jsonCategory)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
