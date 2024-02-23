// Used to show a specific category

package categories

import (
	"encoding/json"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"

	"github.com/go-pg/pg/v10"
)

// Show renders a template to show a given category
func Show(w http.ResponseWriter, r *http.Request) {
	categoryName := r.PathValue("category")
	pageUrl := r.PathValue("pageName")

	if pageUrl == "" && strings.HasSuffix(categoryName, ".json") {
		buildJson(w, r, strings.TrimSuffix(categoryName, ".json"))
		return
	}

	var pullRequests []*models.GithubPullRequest
	category := new(models.Category)
	query := database.DBCon.Model(category).
		Where("category.name = ?", categoryName).
		Relation("PackagesInformation").
		Relation("Packages", func(q *pg.Query) (*pg.Query, error) {
			return q.Order("name ASC"), nil
		})

	pageName := "Packages"
	switch pageUrl {
	case "stabilization":
		pageName = "Stabilization"
		query = query.Relation("Packages.Versions").
			Relation("Packages.Versions.PkgCheckResults", func(q *pg.Query) (*pg.Query, error) {
				return q.Where("class = 'StableRequest'"), nil
			})
	case "outdated":
		pageName = "Outdated"
		query = query.Relation("Packages.Versions").
			Relation("Packages.Outdated")
	case "pull-requests":
		pageName = "Pull requests"
		err := database.DBCon.Model(&pullRequests).
			Join("JOIN package_to_github_pull_requests ON package_to_github_pull_requests.github_pull_request_id = github_pull_request.id").
			Where("package_to_github_pull_requests.package_atom LIKE ?", categoryName+"/%").
			Group("github_pull_request.id").
			Order("github_pull_request.created_at DESC").
			Select()
		if err != nil {
			http.NotFound(w, r)
			return
		}
	case "stabilization.json", "stabilization.xml", "stabilization.list":
		err := query.Relation("Packages.Versions").
			Relation("Packages.Versions.PkgCheckResults", func(q *pg.Query) (*pg.Query, error) {
				return q.Where("class = 'StableRequest'"), nil
			}).Select()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		utils.StabilizationExport(w, pageUrl, category.Packages)
		return
	case "", "packages":
		query = query.Relation("Packages.Versions")
	default:
		http.NotFound(w, r)
		return
	}

	err := query.Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	renderShowPage(w, r, pageName, category, pullRequests)
}

// build the json for the category
func buildJson(w http.ResponseWriter, r *http.Request, categoryName string) {
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
		http.NotFound(w, r)
		return
	}

	jsonCategory.Name = categoryName
	jsonCategory.Href = "https://packages.gentoo.org/categories/" + categoryName

	b, err := json.Marshal(jsonCategory)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
