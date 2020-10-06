package github

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	appUtils "soko/pkg/app/utils"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"soko/pkg/utils"
	"strconv"
	"strings"
	"time"
)

func buildQuery(limit int, isOpen bool, lastUpdated, after string) map[string]string {

	lastUpdatedQuery := ""
	if lastUpdated != "" {
		lastUpdatedQuery = `updated:>` + lastUpdated
	}

	afterQuery := ""
	if after != "" {
		afterQuery = `after: "` + after + `",`
	}

	isOpenQuery := ""
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

	deleteAllPullrequests()

	// year of the git migration
	UpdatePullRequestsAfter(true, "2015-01-01", "")

	updateStatus()
}

func IncrementalUpdatePullRequests() {

	database.Connect()
	defer database.DBCon.Close()

	lastUpdate := appUtils.GetApplicationData().LastUpdate.UTC().Format(time.RFC3339)
	lastUpdate = strings.Split(lastUpdate, "Z")[0] + "Z"
	UpdatePullRequestsAfter(false, lastUpdate, "")
	// TODO --> we need to update old ent
	// TODO --> delete closed pull requests

	updateStatus()
}

func UpdatePullRequestsAfter(isOpen bool, lastUpdated, after string) {
	jsonData := buildQuery(100, isOpen, lastUpdated, after)

	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonValue))
	request.Header.Set("Authorization", "bearer "+config.GithubAPIToken())
	client := &http.Client{Timeout: time.Second * 30}
	response, err := client.Do(request)
	if err != nil {
		logger.Error.Println("The HTTP request failed with error")
		logger.Error.Println(err)
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	var prData models.GitHubPullRequestQueryResult
	err = json.Unmarshal([]byte(data), &prData)

	pullrequests := prData.CreatePullRequest()

	for _, pullrequest := range pullrequests {
		//
		// Create Pullrequest
		//
		database.DBCon.Model(&pullrequest).WherePK().OnConflict("(id) DO UPDATE").Insert()

		//
		// Create Package To Pullrequest
		//
		var affectedPackages []string
		for _, file := range pullrequest.Files {
			pathParts := strings.Split(file.Path, "/")
			if len(pathParts) >= 2 && strings.Contains(pathParts[0], "-") {
				affectedPackages = append(affectedPackages, pathParts[0]+"/"+pathParts[1])
			}
		}
		affectedPackages = utils.Deduplicate(affectedPackages)
		for _, affectedPackage := range affectedPackages {
			database.DBCon.Model(&models.PackageToGithubPullRequest{
				Id:                  affectedPackage + "-" + pullrequest.Id,
				PackageAtom:         affectedPackage,
				GithubPullRequestId: pullrequest.Id,
			}).WherePK().OnConflict("(id) DO UPDATE").Insert()
		}
	}

	//
	// If there is a next page, import it as well
	//

	if prData.HasNextPage() {
		// Wait for some time, as Github will block the request otherwise
		time.Sleep(2 * time.Second)
		UpdatePullRequestsAfter(isOpen, lastUpdated, prData.EndCursor())
	}

}

// deleteAllPullrequests deletes all entries in the pullrequests and package to pull request table
func deleteAllPullrequests() {
	var pullrequests []*models.GithubPullRequest
	database.DBCon.Model(&pullrequests).Select()
	for _, pullrequest := range pullrequests {
		database.DBCon.Model(pullrequest).WherePK().Delete()
	}

	var packagesToGithubPullRequest []*models.PackageToGithubPullRequest
	database.DBCon.Model(&packagesToGithubPullRequest).Select()
	for _, packageToGithubPullRequest := range packagesToGithubPullRequest {
		database.DBCon.Model(packageToGithubPullRequest).WherePK().Delete()
	}
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "pullrequests",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
