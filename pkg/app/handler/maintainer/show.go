package maintainer

import (
	"github.com/go-pg/pg/v9/orm"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
)

// Show renders a template to show a given maintainer page
func Show(w http.ResponseWriter, r *http.Request) {
	maintainerEmail := r.URL.Path[len("/maintainer/"):]
	maintainerEmail = strings.Split(maintainerEmail, "/")[0]
	if !strings.Contains(maintainerEmail, "@") {
		maintainerEmail = maintainerEmail + "@gentoo.org"
	}

	whereClause := "maintainers @> '[{\"Email\": \"" + maintainerEmail + "\"}]'"
	if maintainerEmail == "maintainer-needed@gentoo.org" {
		whereClause = "maintainers IS null"
	}

	maintainer := models.Maintainer{
		Email: maintainerEmail,
	}
	database.DBCon.Model(&maintainer).WherePK().Select()

	var gpackages []*models.Package
	query := database.DBCon.Model(&gpackages).
		Where(whereClause)

	pageName := "packages"
	if strings.HasSuffix(r.URL.Path, "/changelog") {
		pageName = "changelog"
		query = query.
			Relation("Versions").
			Relation("Commits", func(q *orm.Query) (*orm.Query, error) {
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
	} else if strings.HasSuffix(r.URL.Path, "/bugs") {
		pageName = "bugs"
		query = query.
			Relation("Versions").
			Relation("Bugs")
	} else if strings.HasSuffix(r.URL.Path, "/security") {
		pageName = "security"
		query = query.
			Relation("Versions").
			Relation("Bugs")
	} else {
		query = query.
			Relation("Versions").
			Relation("Versions.Masks")
	}

	err := query.Select()

	if err != nil || len(gpackages) == 0 {
		http.NotFound(w, r)
		return
	}

	renderMaintainerTemplate("show",
		"*",
		GetFuncMap(),
		createMaintainerData(pageName, &maintainer, gpackages),
		w)

}
