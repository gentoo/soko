// SPDX-License-Identifier: GPL-2.0-only
package pullrequests

import (
	"iter"
	"log/slog"
	"slices"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"soko/pkg/portage/pullrequests/codeberg"
	"soko/pkg/portage/pullrequests/github"
	"strings"
	"time"
)

func FullUpdatePullRequests() {
	database.Connect()
	defer database.DBCon.Close()

	database.TruncateTable((*models.PullRequest)(nil))
	database.TruncateTable((*models.PackageToPullRequest)(nil))

	updatePullRequests()

	updateStatus()
}

var fetchers = [...]func() iter.Seq[models.PullRequestProvider]{
	codeberg.FetchPullRequests,
	github.FetchPullRequests,
}

func updatePullRequests() {
	categoriesPullRequests := make(map[string]map[string]struct{})
	pullRequestsRows := make([]*models.PullRequest, 0, 1_000)
	var pkgsPullRequests []*models.PackageToPullRequest

	for _, fetcher := range fetchers {
		for pullRequest := range fetcher() {
			pullRequestObject := pullRequest.ToPullRequest()
			pullRequestsRows = append(pullRequestsRows, pullRequestObject)

			affectedPackages := make(map[string]struct{})
			for file := range pullRequest.GetFiles() {
				pathParts := strings.Split(file, "/")
				if len(pathParts) >= 2 && strings.Contains(pathParts[0], "-") {
					affectedPackages[pathParts[0]+"/"+pathParts[1]] = struct{}{}

					prs, ok := categoriesPullRequests[pathParts[0]]
					if !ok {
						prs = make(map[string]struct{})
					}
					prs[pullRequestObject.Id] = struct{}{}
					categoriesPullRequests[pathParts[0]] = prs
				}
			}
			pkgsPullRequests = slices.Grow(pkgsPullRequests, len(affectedPackages))
			for affectedPackage := range affectedPackages {
				pkgsPullRequests = append(pkgsPullRequests, &models.PackageToPullRequest{
					Id:            affectedPackage + "-" + pullRequestObject.Id,
					PackageAtom:   affectedPackage,
					PullRequestId: pullRequestObject.Id,
				})
			}
		}
	}

	if len(pullRequestsRows) == 0 {
		slog.Info("No pull requests to insert")
		return
	}

	result, err := database.DBCon.Model(&pullRequestsRows).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed to insert pull requests", slog.Any("err", err))
		return
	}
	slog.Info("Inserted pull requests", slog.Int("rows", result.RowsAffected()))

	result, err = database.DBCon.Model(&pkgsPullRequests).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed to insert packages to pull requests", slog.Any("err", err))
		return
	}
	slog.Info("Inserted packages to pull requests", slog.Int("rows", result.RowsAffected()))

	updateCategoriesPullRequests(categoriesPullRequests)
}

func updateCategoriesPullRequests(categoriesPullRequests map[string]map[string]struct{}) {
	var categories []*models.CategoryPackagesInformation
	err := database.DBCon.Model(&categories).Column("name").Select()
	if err != nil {
		slog.Error("Failed fetching categories packages information", slog.Any("err", err))
		return
	} else if len(categories) > 0 {
		for _, category := range categories {
			category.PullRequests = len(categoriesPullRequests[category.Name])
			delete(categoriesPullRequests, category.Name)
		}
		_, err = database.DBCon.Model(&categories).Set("pull_requests = ?pull_requests").Update()
		if err != nil {
			slog.Error("Failed updating categories packages information", slog.Any("err", err))
		}
		categories = make([]*models.CategoryPackagesInformation, 0, len(categoriesPullRequests))
	}

	for category, prs := range categoriesPullRequests {
		categories = append(categories, &models.CategoryPackagesInformation{
			Name:         category,
			PullRequests: len(prs),
		})
	}
	if len(categories) > 0 {
		_, err = database.DBCon.Model(&categories).Insert()
		if err != nil {
			slog.Error("Failed inserting categories packages information", slog.Any("err", err))
		}
	}
}

func updateStatus() {
	_, err := database.DBCon.Model(&models.Application{
		Id:         "pullrequests",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed updating status", slog.Any("err", err))
	}
}
