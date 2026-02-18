// SPDX-License-Identifier: GPL-2.0-only
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"soko/pkg/config"
	"soko/pkg/models"
	"strconv"
	"time"
)

var client = &http.Client{Timeout: time.Second * 30}

func FetchPullRequests() iter.Seq[models.PullRequestProvider] {
	const isOpen = true
	const lastUpdated = "2015-01-01" // year of the git migration

	token := config.GithubAPIToken()

	return func(yield func(models.PullRequestProvider) bool) {
		var after string
		index := 0
		for {
			for limit := 100; limit >= 8; limit /= 2 {
				time.Sleep(2 * time.Second)
				slog.Info("Requesting pull requests", slog.Int("index", index), slog.Int("limit", limit))
				data, statusCode, err := fetchPullRequestsBatch(token, limit, isOpen, lastUpdated, after)
				if err != nil {
					if statusCode == http.StatusGatewayTimeout || statusCode == http.StatusBadGateway {
						slog.Warn("Query too big, reducing from limit", slog.Int("limit", limit))
						continue
					}
					slog.Error("Failed to fetch pull requests", slog.Any("err", err))
					return
				}

				for _, rawObject := range data.Data.Search.Edges {
					if !yield(&rawObject.Node) {
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

func fetchPullRequestsBatch(
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
