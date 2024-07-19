// SPDX-License-Identifier: GPL-2.0-only
package github

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"strconv"
	"strings"
	"time"
)

func buildQuery(limit int, isOpen bool, lastUpdated, after string) map[string]string {
	var lastUpdatedQuery string
	if lastUpdated != "" {
		lastUpdatedQuery = `updated:>` + lastUpdated
	}

	var afterQuery string
	if after != "" {
		afterQuery = `after: "` + after + `",`
	}

	var isOpenQuery string
	if isOpen {
		isOpenQuery = `is:open`
	}

	return map[string]string{
		"query": `
            {
			  rateLimit {
				limit
				cost
				remaining
				resetAt
			  }
			  search(query: "repo:gentoo/gentoo is:pr ` + isOpenQuery + ` ` + lastUpdatedQuery + `", type: ISSUE, ` + afterQuery + ` last: ` + strconv.Itoa(limit) + `) {
				pageInfo {
				  startCursor
				  hasNextPage
				  endCursor
				}
				edges {
				  node {
					... on PullRequest {
					  number
					  closed
					  url
					  title
					  createdAt
					  updatedAt
					  comments {
						totalCount
					  }
					  files(first: 50) {
						edges {
						  node {
							path
							additions
							deletions
						  }
						}
					  }

					  author {
						login
					  }
					  commits(last: 1){
					  nodes{
						commit{
						  commitUrl
						  oid
						  status {
							state

							contexts {
							  state
							  targetUrl
							  description
							  context
							}
						  }
						}
					  }
					}
					  labels(first:10) {
						edges {
						  node {
							name
							color
						  }
						}
					  }
					}
				  }
				}
			  }
			}
        `,
	}
}

func FullUpdatePullRequests() {
	database.Connect()
	defer database.DBCon.Close()

	database.TruncateTable((*models.GithubPullRequest)(nil))
	database.TruncateTable((*models.PackageToGithubPullRequest)(nil))

	// year of the git migration
	updatePullRequestsAfter(true, "2015-01-01", "")

	updateStatus()
}

func updatePullRequestsAfter(isOpen bool, lastUpdated, after string) {
	pullRequests := make(map[int]*models.GithubPullRequest)
	client := &http.Client{Timeout: time.Second * 30}

	for {
		slog.Info("Requesting pull requests", slog.Int("index", len(pullRequests)))
		jsonData := buildQuery(100, isOpen, lastUpdated, after)
		jsonValue, _ := json.Marshal(jsonData)

		request, err := http.NewRequest(http.MethodPost, "https://api.github.com/graphql", bytes.NewBuffer(jsonValue))
		if err != nil {
			slog.Error("Failed querying github graphql", slog.Any("err", err))
			return
		}

		request.Header.Set("Authorization", "bearer "+config.GithubAPIToken())
		response, err := client.Do(request)
		if err != nil {
			slog.Error("The HTTP request failed", slog.Any("err", err))
			return
		}
		defer response.Body.Close()

		var prData models.GitHubPullRequestQueryResult
		err = json.NewDecoder(response.Body).Decode(&prData)
		if err != nil {
			slog.Error("Failed to parse JSON", slog.Any("err", err))
			return
		}
		prData.AppendPullRequest(pullRequests)

		// If there is a next page, import it as well
		if prData.HasNextPage() {
			time.Sleep(2 * time.Second)
			after = prData.EndCursor()
		} else {
			break
		}
	}

	if len(pullRequests) == 0 {
		slog.Info("No pull requests to insert")
		return
	}

	categoriesPullRequests := make(map[string]map[string]struct{})
	var pkgsPullRequests []*models.PackageToGithubPullRequest
	for _, pullrequest := range pullRequests {
		affectedPackages := make(map[string]struct{})
		for _, file := range pullrequest.Files {
			pathParts := strings.Split(file.Path, "/")
			if len(pathParts) >= 2 && strings.Contains(pathParts[0], "-") {
				affectedPackages[pathParts[0]+"/"+pathParts[1]] = struct{}{}

				prs, ok := categoriesPullRequests[pathParts[0]]
				if !ok {
					prs = make(map[string]struct{})
				}
				prs[pullrequest.Id] = struct{}{}
				categoriesPullRequests[pathParts[0]] = prs
			}
		}
		for affectedPackage := range affectedPackages {
			pkgsPullRequests = append(pkgsPullRequests, &models.PackageToGithubPullRequest{
				Id:                  affectedPackage + "-" + pullrequest.Id,
				PackageAtom:         affectedPackage,
				GithubPullRequestId: pullrequest.Id,
			})
		}
	}

	rows := make([]*models.GithubPullRequest, 0, len(pullRequests))
	for _, row := range pullRequests {
		rows = append(rows, row)
	}
	result, err := database.DBCon.Model(&rows).OnConflict("(id) DO UPDATE").Insert()
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
	database.DBCon.Model(&models.Application{
		Id:         "pullrequests",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
