package github

import (
	"bytes"
	"encoding/json"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
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

	database.TruncateTable[models.GithubPullRequest]("id")
	database.TruncateTable[models.PackageToGithubPullRequest]("id")

	// year of the git migration
	UpdatePullRequestsAfter(true, "2015-01-01", "")

	updateStatus()
}

func IncrementalUpdatePullRequests() {

	database.Connect()
	defer database.DBCon.Close()

	lastUpdate := utils.GetApplicationData().LastUpdate.UTC().Format(time.RFC3339)
	lastUpdate = strings.Split(lastUpdate, "Z")[0] + "Z"
	UpdatePullRequestsAfter(false, lastUpdate, "")
	// TODO --> we need to update old ent
	// TODO --> delete closed pull requests

	updateStatus()
}

func UpdatePullRequestsAfter(isOpen bool, lastUpdated, after string) {
	pullRequests := make(map[int]*models.GithubPullRequest)
	client := &http.Client{Timeout: time.Second * 30}

	for {
		logger.Info.Println("Requesting pull requests starting with", len(pullRequests))
		jsonData := buildQuery(100, isOpen, lastUpdated, after)
		jsonValue, _ := json.Marshal(jsonData)

		request, err := http.NewRequest(http.MethodPost, "https://api.github.com/graphql", bytes.NewBuffer(jsonValue))
		if err != nil {
			logger.Error.Println("Failed to query github graphql", err)
			return
		}

		request.Header.Set("Authorization", "bearer "+config.GithubAPIToken())
		response, err := client.Do(request)
		if err != nil {
			logger.Error.Println("The HTTP request failed with error", err)
			return
		}
		defer response.Body.Close()

		var prData models.GitHubPullRequestQueryResult
		err = json.NewDecoder(response.Body).Decode(&prData)
		if err != nil {
			logger.Error.Println("Failed to parse JSON", err)
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
		logger.Info.Println("No pull requests to insert")
		return
	}

	var pkgsPullRequests []*models.PackageToGithubPullRequest
	for _, pullrequest := range pullRequests {
		affectedPackages := make(map[string]struct{})
		for _, file := range pullrequest.Files {
			pathParts := strings.Split(file.Path, "/")
			if len(pathParts) >= 2 && strings.Contains(pathParts[0], "-") {
				affectedPackages[pathParts[0]+"/"+pathParts[1]] = struct{}{}
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
		logger.Error.Println("Failed to insert pull requests", err)
		return
	}
	logger.Info.Println("Inserted", result.RowsAffected(), "pull requests")

	result, err = database.DBCon.Model(&pkgsPullRequests).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		logger.Error.Println("Failed to insert packages to pull requests", err)
		return
	}
	logger.Info.Println("Inserted", result.RowsAffected(), "packages to pull requests")
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "pullrequests",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
