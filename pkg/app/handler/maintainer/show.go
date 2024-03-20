package maintainer

import (
	"encoding/json"
	"net/http"
	"soko/pkg/app/handler/packages/components"
	"soko/pkg/app/layout"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/gorilla/feeds"
)

func common(w http.ResponseWriter, r *http.Request) (maintainer models.Maintainer, packagesQuery *pg.Query, packagesCount int, err error) {
	maintainerEmail := r.PathValue("email")
	if !strings.Contains(maintainerEmail, "@") {
		maintainerEmail += "@gentoo.org"
	}

	packagesQuery = database.DBCon.Model((*models.Package)(nil)).Column("atom")

	if maintainerEmail == "maintainer-needed@gentoo.org" {
		packagesQuery = packagesQuery.Where("NULLIF(maintainers, '[]') IS null")
	} else {
		packagesQuery = packagesQuery.Where("maintainers @> ?", `[{"Email": "`+maintainerEmail+`"}]`)
	}

	maintainer = models.Maintainer{Email: maintainerEmail}
	err = database.DBCon.Model(&maintainer).WherePK().Relation("Project").Relation("Projects").Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	userPreferences := utils.GetUserPreferences(r)
	if userPreferences.Maintainers.IncludeProjectPackages && maintainer.Projects != nil && len(maintainer.Projects) > 0 {
		excludeList := strings.Join(userPreferences.Maintainers.ExcludedProjects, ",")
		for _, proj := range maintainer.Projects {
			if !strings.Contains(excludeList, proj.Email) {
				packagesQuery = packagesQuery.WhereOr("maintainers @> ?", `[{"Email": "`+proj.Email+`"}]`)
			}
		}
	}

	packagesCount, err = packagesQuery.Clone().Count()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	return
}

func ShowChangelog(w http.ResponseWriter, r *http.Request) {
	maintainer, query, packagesCount, err := common(w, r)
	if err != nil {
		return
	}
	var commits []*models.Commit
	err = database.DBCon.Model(&commits).
		Join("JOIN commit_to_packages").JoinOn("commit.id = commit_to_packages.commit_id").
		Where("commit_to_packages.package_atom IN (?)", query).
		Order("preceding_commits DESC").
		Limit(50).
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	layout.Layout(maintainer.Name, "maintainers",
		show(packagesCount, &maintainer, "Changelog", components.Changelog("", commits)),
	).Render(r.Context(), w)
}

func ShowChangelogFeed(w http.ResponseWriter, r *http.Request) {
	maintainer, query, _, err := common(w, r)
	if err != nil {
		return
	}
	var commits []*models.Commit
	err = database.DBCon.Model(&commits).
		Join("JOIN commit_to_packages").JoinOn("commit.id = commit_to_packages.commit_id").
		Where("commit_to_packages.package_atom IN (?)", query).
		Order("preceding_commits DESC").
		Limit(100).
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	feed := &feeds.Feed{
		Title:       "100 latest commits for " + maintainer.Name,
		Description: "100 latest commits for " + maintainer.Name,
		Author:      &feeds.Author{Name: "Gentoo Packages Database"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: "https://packages.gentoo.org/maintainer/" + maintainer.Email + "/changelog"},
	}

	for _, commit := range commits {
		feed.Add(&feeds.Item{
			Title:   commit.Message,
			Updated: commit.CommitterDate,
			Created: commit.AuthorDate,
			Author:  &feeds.Author{Name: commit.CommitterName, Email: commit.CommitterEmail},
			Link:    &feeds.Link{Href: "https://gitweb.gentoo.org/repo/gentoo.git/commit/?id=" + commit.Id, Type: "text/html", Rel: "alternate"},
			Id:      commit.Id,
		})
	}
	feed.WriteAtom(w)
}

func ShowOutdated(w http.ResponseWriter, r *http.Request) {
	maintainer, query, packagesCount, err := common(w, r)
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
		Where("atom IN (?)", query).
		Order("atom").
		Select(&outdated)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	layout.Layout(maintainer.Name, "maintainers",
		show(packagesCount, &maintainer, "Outdated", components.Outdated(outdated)),
	).Render(r.Context(), w)
}

func ShowOutdatedFeed(w http.ResponseWriter, r *http.Request) {
	maintainer, query, _, err := common(w, r)
	if err != nil {
		return
	}
	var outdated []models.OutdatedPackages
	err = database.DBCon.Model(&outdated).
		Where("atom IN (?)", query).
		Order("atom").
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	utils.OutdatedFeed(w, "https://packages.gentoo.org/maintainer/"+maintainer.Email+"/outdated", maintainer.Name+" <"+maintainer.Email+">", outdated)
}

func ShowPullRequests(w http.ResponseWriter, r *http.Request) {
	maintainer, query, packagesCount, err := common(w, r)
	if err != nil {
		return
	}
	var pullRequests []*models.GithubPullRequest
	err = database.DBCon.Model(&pullRequests).
		DistinctOn("github_pull_request.id").
		OrderExpr("github_pull_request.id DESC").
		Join("JOIN package_to_github_pull_requests").JoinOn("github_pull_request.id = package_to_github_pull_requests.github_pull_request_id").
		Where("package_atom IN (?)", query).
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	layout.Layout(maintainer.Name, "maintainers",
		show(packagesCount, &maintainer, "Pull requests", components.PullRequests(len(pullRequests) > 0, pullRequests)),
	).Render(r.Context(), w)
}

func ShowStabilization(w http.ResponseWriter, r *http.Request) {
	maintainer, query, packagesCount, err := common(w, r)
	if err != nil {
		return
	}
	var results []*models.PkgCheckResult
	err = database.DBCon.Model(&results).
		Column("atom", "cpv", "message").
		Where("class = ?", "StableRequest").
		Where("atom IN (?)", query).
		OrderExpr("cpv").
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	layout.Layout(maintainer.Name, "maintainers",
		show(packagesCount, &maintainer, "Stabilization", components.Stabilizations(len(results) > 0, results)),
	).Render(r.Context(), w)
}

func ShowBugs(w http.ResponseWriter, r *http.Request) {
	maintainer, query, packagesCount, err := common(w, r)
	if err != nil {
		return
	}
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	generalCount, stabilizationCount, keywordingCount := countBugsCategories(bugs)
	layout.Layout(maintainer.Name, "maintainers",
		show(packagesCount, &maintainer, "Bugs", components.Bugs("", generalCount, stabilizationCount, keywordingCount, bugs)),
	).Render(r.Context(), w)
}

func ShowSecurity(w http.ResponseWriter, r *http.Request) {
	maintainer, query, packagesCount, err := common(w, r)
	if err != nil {
		return
	}
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	layout.Layout(maintainer.Name, "maintainers",
		show(packagesCount, &maintainer, "Security", components.SecurityBugs("", len(bugs) > 0, bugs)),
	).Render(r.Context(), w)
}

func ShowStabilizationFile(w http.ResponseWriter, r *http.Request) {
	_, query, _, err := common(w, r)
	if err != nil {
		return
	}
	pageName := r.URL.Path[strings.LastIndexByte(r.URL.Path, '/')+1:]

	var results []*models.PkgCheckResult
	err = database.DBCon.Model(&results).
		Column("atom", "cpv", "message").
		Where("class = ?", "StableRequest").
		Where("atom IN (?)", query).
		OrderExpr("cpv").
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	utils.StabilizationExport(w, pageName, results)
}

func ShowStabilizationFeed(w http.ResponseWriter, r *http.Request) {
	maintainer, query, _, err := common(w, r)
	if err != nil {
		return
	}

	var results []*models.PkgCheckResult
	err = database.DBCon.Model(&results).
		Column("atom", "cpv", "message").
		Where("class = ?", "StableRequest").
		Where("atom IN (?)", query).
		OrderExpr("cpv").
		Select()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	utils.StabilizationFeed(w, "https://packages.gentoo.org/maintainer/"+maintainer.Email+"/stabilization", maintainer.Name+" <"+maintainer.Email+">", results)
}

func ShowPackages(w http.ResponseWriter, r *http.Request) {
	maintainer, query, packagesCount, err := common(w, r)
	if err != nil {
		return
	}
	var gpackages []*models.Package
	err = query.Model(&gpackages).
		Column("category").
		Order("category", "name").
		Relation("Versions").Select()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	layout.Layout(maintainer.Name, "maintainers",
		show(packagesCount, &maintainer, "Packages", showPackages(gpackages, &maintainer)),
	).Render(r.Context(), w)
}

func ShowInfoJson(w http.ResponseWriter, r *http.Request) {
	maintainer, _, _, err := common(w, r)
	if err != nil {
		return
	}

	var reply struct {
		Email     string   `json:"email"`
		Name      string   `json:"name"`
		IsProject bool     `json:"is_project"`
		Members   []string `json:"members"`
		MemberOf  []string `json:"member_of"`
	}

	reply.Email = maintainer.Email
	reply.Name = maintainer.Name
	reply.IsProject = maintainer.Type == "project"

	for _, member := range maintainer.Project.Members {
		reply.Members = append(reply.Members, member.Email)
	}
	for _, project := range maintainer.Projects {
		reply.MemberOf = append(reply.MemberOf, project.Email)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(reply)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
