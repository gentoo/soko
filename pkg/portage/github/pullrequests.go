// SPDX-License-Identifier: GPL-2.0-only
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"maps"
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

	updatePullRequestsAfter()

	updateStatus()
}

var client = &http.Client{Timeout: time.Second * 30}

func fetchPullRequests(
	token string, limit int, isOpen bool, lastUpdated, after string,
) (data GitHubPullRequestQueryResult, statusCode int, err error) {
	jsonData := buildQuery(limit, isOpen, lastUpdated, after)
	jsonValue, _ := json.Marshal(jsonData)

	request, err := http.NewRequest(http.MethodPost, "https://api.github.com/graphql", bytes.NewBuffer(jsonValue))
	if err != nil {
		slog.Error("Failed querying github graphql", slog.Any("err", err))
		return
	}

	request.Header.Set("Authorization", "bearer "+token)
	response, err := client.Do(request)
	if err != nil {
		slog.Error("The HTTP request failed", slog.Any("err", err))
		return
	}
	defer response.Body.Close()

	statusCode = response.StatusCode
	if statusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		if len(body) > 100 {
			body = body[:100]
		}
		slog.Error("The HTTP request failed with status code", slog.Int("status", response.StatusCode), slog.String("body", string(body)))
		err = fmt.Errorf("status code: %d", statusCode)
		return
	}

	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		slog.Error("Failed to parse JSON", slog.Any("err", err))
		return
	}
	return
}

func fetchAllPullRequests() iter.Seq2[int, *models.GithubPullRequest] {
	const isOpen = true
	const lastUpdated = "2015-01-01" // year of the git migration

	token := config.GithubAPIToken()

	return func(yield func(int, *models.GithubPullRequest) bool) {
		var after string
		index := 0
		for {
			for limit := 100; limit >= 8; limit /= 2 {
				time.Sleep(2 * time.Second)
				slog.Info("Requesting pull requests", slog.Int("index", index), slog.Int("limit", limit))
				data, statusCode, err := fetchPullRequests(token, limit, isOpen, lastUpdated, after)
				if err != nil {
					if statusCode == http.StatusGatewayTimeout || statusCode == http.StatusBadGateway {
						slog.Warn("Query too big, reducing from limit", slog.Int("limit", limit))
						continue
					}
					slog.Error("Failed to fetch pull requests", slog.Any("err", err))
					return
				}

				for _, rawObject := range data.Data.Search.Edges {
					pullRequest := rawObject.Node
					var ciState, ciStateLink string
					if nodes := pullRequest.Commits.Nodes; len(nodes) > 0 {
						ciState = nodes[0].Commit.Status.State

						if contexts := nodes[0].Commit.Status.Contexts; len(contexts) > 0 {
							ciStateLink = contexts[0].TargetUrl
						}
					}

					if !yield(pullRequest.Number, &models.GithubPullRequest{
						Id:          strconv.Itoa(pullRequest.Number),
						Closed:      pullRequest.Closed,
						Url:         pullRequest.Url,
						Title:       pullRequest.Title,
						CreatedAt:   pullRequest.CreatedAt,
						UpdatedAt:   pullRequest.UpdatedAt,
						CiState:     ciState,
						CiStateLink: ciStateLink,
						Labels:      pullRequest.CreateLabelsArray(),
						Comments:    pullRequest.Comments.TotalCount,
						Files:       pullRequest.CreateFilesArray(),
						Author:      pullRequest.Author.Login,
					}) {
						return
					}
				}
				index += len(data.Data.Search.Edges)

				if !data.HasNextPage() {
					return // Finished
				}
				after = data.EndCursor()
				break
			}
		}
	}
}

func updatePullRequestsAfter() {
	pullRequests := make(map[int]*models.GithubPullRequest, 1_000)
	maps.Insert(pullRequests, fetchAllPullRequests())

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
	_, err := database.DBCon.Model(&models.Application{
		Id:         "pullrequests",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed updating status", slog.Any("err", err))
	}
}
