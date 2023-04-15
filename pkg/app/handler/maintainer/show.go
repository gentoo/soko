package maintainer

import (
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"

	"github.com/go-pg/pg/v10"
)

// Show renders a template to show a given maintainer page
func Show(w http.ResponseWriter, r *http.Request) {
	maintainerEmail, pageName, _ := strings.Cut(r.URL.Path[len("/maintainer/"):], "/")
	if !strings.Contains(maintainerEmail, "@") {
		maintainerEmail += "@gentoo.org"
	}

	var gpackages []*models.Package
	query := database.DBCon.Model(&gpackages).
		Order("category", "name")

	if maintainerEmail == "maintainer-needed@gentoo.org" {
		query = query.Where("NULLIF(maintainers, '[]') IS null")
	} else {
		query = query.Where("maintainers @> ?", `[{"Email": "`+maintainerEmail+`"}]`)
	}

	maintainer := models.Maintainer{
		Email: maintainerEmail,
	}
	err := database.DBCon.Model(&maintainer).WherePK().Relation("Project").Relation("Projects").Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	userPreferences := utils.GetUserPreferences(r)
	if userPreferences.Maintainers.IncludeProjectPackages && maintainer.Projects != nil && len(maintainer.Projects) > 0 {
		excludeList := strings.Join(userPreferences.Maintainers.ExcludedProjects, ",")
		for _, proj := range maintainer.Projects {
			if !strings.Contains(excludeList, proj.Email) {
				query = query.WhereOr("maintainers @> ?", `[{"Email": "`+proj.Email+`"}]`)
			}
		}
	}

	switch pageName {
	case "changelog":
		query = query.
			Relation("Versions").
			Relation("Commits", func(q *pg.Query) (*pg.Query, error) {
				return q.Order("preceding_commits DESC").Limit(50), nil
			})
	case "outdated":
		query = query.
			Relation("Versions").
			Relation("Outdated")
	case "qa-report":
		query = query.
			Relation("Versions").
			Relation("PkgCheckResults").
			Relation("Versions.PkgCheckResults")
	case "pull-requests":
		query = query.
			Relation("Versions").
			Relation("PullRequests")
	case "stabilization":
		query = query.
			Relation("Versions").
			Relation("Versions.PkgCheckResults", func(q *pg.Query) (*pg.Query, error) {
				return q.Where("class = ?", "StableRequest"), nil
			}).
			Relation("Bugs")
	case "bugs":
		query = query.
			Relation("Versions").
			Relation("Versions.Bugs").
			Relation("Bugs")
	case "security":
		query = query.
			Relation("Versions").
			Relation("Versions.Bugs").
			Relation("Bugs")
	case "stabilization.json", "stabilization.xml", "stabilization.list":
		err = query.
			Relation("Versions").
			Relation("Versions.PkgCheckResults", func(q *pg.Query) (*pg.Query, error) {
				return q.Where("class = ?", "StableRequest"), nil
			}).Select()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		utils.StabilizationExport(w, pageName, gpackages)
		return
	default:
		pageName = "packages"
		query = query.
			Relation("Versions").
			Relation("Versions.Masks")
	}

	err = query.Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	renderMaintainerTemplate("show",
		"*",
		GetFuncMap(),
		createMaintainerData(pageName, &maintainer, gpackages),
		w)
}
