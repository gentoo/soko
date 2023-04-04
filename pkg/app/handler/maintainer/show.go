package maintainer

import (
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	"sort"
	"strings"

	"github.com/go-pg/pg/v10"
)

// Show renders a template to show a given maintainer page
func Show(w http.ResponseWriter, r *http.Request) {
	maintainerEmail := r.URL.Path[len("/maintainer/"):]
	maintainerEmail, _, _ = strings.Cut(maintainerEmail, "/")
	if !strings.Contains(maintainerEmail, "@") {
		maintainerEmail = maintainerEmail + "@gentoo.org"
	}

	var gpackages []*models.Package
	query := database.DBCon.Model(&gpackages)

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

	pageName := "packages"
	if strings.HasSuffix(r.URL.Path, "/changelog") {
		pageName = "changelog"
		query = query.
			Relation("Versions").
			Relation("Commits", func(q *pg.Query) (*pg.Query, error) {
				return q.Order("preceding_commits DESC").Limit(50), nil
			})
	} else if strings.HasSuffix(r.URL.Path, "/outdated") {
		pageName = "outdated"
		query = query.
			Relation("Versions").
			Relation("Outdated")
	} else if strings.HasSuffix(r.URL.Path, "/qa-report") {
		pageName = "qa-report"
		query = query.
			Relation("Versions").
			Relation("PkgCheckResults").
			Relation("Versions.PkgCheckResults")
	} else if strings.HasSuffix(r.URL.Path, "/pull-requests") {
		pageName = "pull-requests"
		query = query.
			Relation("Versions").
			Relation("PullRequests")
	} else if strings.HasSuffix(r.URL.Path, "/stabilization") {
		pageName = "stabilization"
		query = query.
			Relation("Versions").
			Relation("PkgCheckResults").
			Relation("Versions.PkgCheckResults").
			Relation("Bugs")
	} else if strings.HasSuffix(r.URL.Path, "/bugs") {
		pageName = "bugs"
		query = query.
			Relation("Versions").
			Relation("Versions.Bugs").
			Relation("Bugs")
	} else if strings.HasSuffix(r.URL.Path, "/security") {
		pageName = "security"
		query = query.
			Relation("Versions").
			Relation("Versions.Bugs").
			Relation("Bugs")
	} else {
		query = query.
			Relation("Versions").
			Relation("Versions.Masks")
	}

	err = query.Select()

	if err != nil {
		http.NotFound(w, r)
		return
	}

	sort.Slice(gpackages, func(i, j int) bool {
		if gpackages[i].Category != gpackages[j].Category {
			return gpackages[i].Category < gpackages[j].Category
		}
		return gpackages[i].Name < gpackages[j].Name
	})

	renderMaintainerTemplate("show",
		"*",
		GetFuncMap(),
		createMaintainerData(pageName, &maintainer, gpackages),
		w)

}
