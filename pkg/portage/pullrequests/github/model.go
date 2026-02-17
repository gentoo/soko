// SPDX-License-Identifier: GPL-2.0-only
package github

import (
	"iter"
	"soko/pkg/models"
	"strconv"
)

type GitHubPullRequestQueryResult struct {
	Data GitHubPullRequestQueryResultData `json:"data"`
}

func (res *GitHubPullRequestQueryResult) HasNextPage() bool {
	return res.Data.Search.PageInfo.HasNextPage
}

func (res *GitHubPullRequestQueryResult) EndCursor() string {
	return res.Data.Search.PageInfo.EndCursor
}

type GitHubPullRequestQueryResultData struct {
	Search GitHubPullRequestSearchResult `json:"search"`
}

type GitHubPullRequestSearchResult struct {
	PageInfo GitHubPullRequestSearchPageInfo `json:"pageInfo"`
	Edges    []GitHubPullRequestSearchEdge   `json:"edges"`
}

type GitHubPullRequestSearchPageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	StartCursor string `json:"startCursor"`
	EndCursor   string `json:"endCursor"`
}

type GitHubPullRequestSearchEdge struct {
	Node GitHubPullRequestSearchNode `json:"node"`
}

type GitHubPullRequestSearchNode struct {
	Number    int                       `json:"number"`
	Closed    bool                      `json:"closed"`
	Url       string                    `json:"url"`
	Title     string                    `json:"title"`
	CreatedAt string                    `json:"createdAt"`
	UpdatedAt string                    `json:"updatedAt"`
	Comments  GitHubPullRequestComments `json:"comments"`
	Files     GitHubPullRequestFiles    `json:"files"`
	Author    GitHubPullRequestAuthor   `json:"author"`
	Labels    GitHubPullRequestLabels   `json:"labels"`
	Commits   GitHubPullRequestCommits  `json:"commits"`
}

func (pr *GitHubPullRequestSearchNode) ToPullRequest() *models.PullRequest {
	var ciState, ciStateLink string
	if nodes := pr.Commits.Nodes; len(nodes) > 0 {
		ciState = nodes[0].Commit.Status.State

		if contexts := nodes[0].Commit.Status.Contexts; len(contexts) > 0 {
			ciStateLink = contexts[0].TargetUrl
		}
	}

	labels := make([]models.PullRequestLabel, len(pr.Labels.Edges))
	for i, label := range pr.Labels.Edges {
		labels[i] = models.PullRequestLabel{Name: label.Node.Name, Color: label.Node.Color}
	}

	return &models.PullRequest{
		Id:          "github/" + strconv.Itoa(pr.Number),
		Closed:      pr.Closed,
		Url:         pr.Url,
		Title:       pr.Title,
		CreatedAt:   pr.CreatedAt,
		UpdatedAt:   pr.UpdatedAt,
		CiState:     ciState,
		CiStateLink: ciStateLink,
		Labels:      labels,
		Comments:    pr.Comments.TotalCount,
		Author:      pr.Author.Login,
	}
}

func (pr *GitHubPullRequestSearchNode) GetFiles() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, file := range pr.Files.Edges {
			if !yield(file.Node.Path) {
				return
			}
		}
	}
}

type GitHubPullRequestCommits struct {
	Nodes []GitHubPullRequestCommitNode `json:"nodes"`
}

type GitHubPullRequestCommitNode struct {
	Commit GitHubPullRequestCommit `json:"commit"`
}

type GitHubPullRequestCommit struct {
	CommitUrl string                        `json:"commitUrl"`
	Oid       string                        `json:"oid"`
	Status    GitHubPullRequestCommitStatus `json:"status"`
}

type GitHubPullRequestCommitStatus struct {
	State    string                                 `json:"state"`
	Contexts []GitHubPullRequestCommitStatusContext `json:"contexts"`
}

type GitHubPullRequestCommitStatusContext struct {
	State       string `json:"state"`
	TargetUrl   string `json:"targetUrl"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

type GitHubPullRequestLabels struct {
	Edges []GitHubPullRequestLabelEdge `json:"edges"`
}

type GitHubPullRequestLabelEdge struct {
	Node GitHubPullRequestLabelNode `json:"node"`
}

type GitHubPullRequestLabelNode struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type GitHubPullRequestAuthor struct {
	Login string `json:"login"`
}

type GitHubPullRequestFiles struct {
	Edges []GitHubPullRequestFileEdge `json:"edges"`
}

type GitHubPullRequestFileEdge struct {
	Node GitHubPullRequestFileNode `json:"node"`
}

type GitHubPullRequestFileNode struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

type GitHubPullRequestComments struct {
	TotalCount int `json:"totalCount"`
}
