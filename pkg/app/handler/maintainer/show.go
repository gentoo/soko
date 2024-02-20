package maintainer

import (
	"net/http"
	"soko/pkg/app/handler/packages/components"
	"soko/pkg/app/layout"
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
		Column("atom")

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

	packagesCount, err := query.Clone().Count()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch pageName {
	case "changelog":
		var commits []*models.Commit
		err = database.DBCon.Model(&commits).
			Join("JOIN commit_to_packages").JoinOn("commit.id = commit_to_packages.commit_id").
			Where("commit_to_packages.package_atom IN (?)", query).
			Order("preceding_commits DESC").
			Limit(50).
			Select()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		layout.Layout(maintainer.Name, "maintainers",
			show(packagesCount, &maintainer, "Changelog", components.Changelog("", commits)),
		).Render(r.Context(), w)
		return
	case "outdated":
		var outdated []components.OutdatedItem
		descriptionQuery := database.DBCon.Model((*models.Version)(nil)).
			Column("description").
			Where("atom = outdated_packages.atom").
			Limit(1)
		err := database.DBCon.Model((*models.OutdatedPackages)(nil)).
			Column("atom").ColumnExpr("(?) AS description", descriptionQuery).
			Where("atom IN (?)", query).
			Order("atom").
			Select(&outdated)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		layout.Layout(maintainer.Name, "maintainers",
			show(packagesCount, &maintainer, "Outdated", components.Outdated(outdated)),
		).Render(r.Context(), w)
		return
	case "pull-requests":
		var pullRequests []models.GithubPullRequest
		err = database.DBCon.Model(&pullRequests).
			DistinctOn("github_pull_request.id").
			OrderExpr("github_pull_request.id DESC").
			Join("JOIN package_to_github_pull_requests").JoinOn("github_pull_request.id = package_to_github_pull_requests.github_pull_request_id").
			Where("package_atom IN (?)", query).
			Select()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		layout.Layout(maintainer.Name, "maintainers",
			show(packagesCount, &maintainer, "Pull requests", components.PullRequests(len(pullRequests) > 0, pullRequests)),
		).Render(r.Context(), w)
		return
	case "stabilization":
		var results []*models.PkgCheckResult
		err = database.DBCon.Model(&results).
			Column("atom", "cpv", "message").
			Where("class = ?", "StableRequest").
			Where("atom IN (?)", query).
			OrderExpr("cpv").
			Select()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		layout.Layout(maintainer.Name, "maintainers",
			show(packagesCount, &maintainer, "Stabilization", components.Stabilizations(len(results) > 0, results)),
		).Render(r.Context(), w)
		return
	case "bugs":
		var bugs []*models.Bug
		err = database.DBCon.Model(&bugs).
			DistinctOn("id::INT").
			Column("id", "summary", "component", "assignee").
			OrderExpr("id::INT").
			With("wanted", query).
			Where("id IN (?)",
				database.DBCon.Model((*models.PackageToBug)(nil)).
					Column("bug_id").
					Join("JOIN wanted").JoinOn("package_atom = wanted.atom")).
			WhereOr("id IN (?)",
				database.DBCon.Model((*models.VersionToBug)(nil)).
					Column("bug_id").
					Join("JOIN versions").JoinOn("version_id = versions.id").
					Join("JOIN wanted").JoinOn("versions.atom = wanted.atom")).
			Select()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		generalCount, stabilizationCount, keywordingCount := countBugsCategories(bugs)
		layout.Layout(maintainer.Name, "maintainers",
			show(packagesCount, &maintainer, "Bugs", components.Bugs(generalCount, stabilizationCount, keywordingCount, bugs)),
		).Render(r.Context(), w)
		return
	case "security":
		var bugs []*models.Bug
		err = database.DBCon.Model(&bugs).
			DistinctOn("id::INT").
			Column("id", "summary", "component", "assignee").
			OrderExpr("id::INT").
			With("wanted", query).
			Where("component = ?", "Vulnerabilities").
			Where("id IN (?)",
				database.DBCon.Model((*models.PackageToBug)(nil)).
					Column("bug_id").
					Join("JOIN wanted").JoinOn("package_atom = wanted.atom")).
			Select()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		layout.Layout(maintainer.Name, "maintainers",
			show(packagesCount, &maintainer, "Security", components.SecurityBugs(len(bugs) > 0, bugs)),
		).Render(r.Context(), w)
		return
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
		err = query.Column("category").
			Order("category", "name").
			Relation("Versions").Select()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		layout.Layout(maintainer.Name, "maintainers",
			show(packagesCount, &maintainer, "Packages", showPackages(gpackages, &maintainer)),
		).Render(r.Context(), w)
		return
	}
}
