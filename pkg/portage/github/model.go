// SPDX-License-Identifier: GPL-2.0-only
package github

import (
	"soko/pkg/models"
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

func (node *GitHubPullRequestSearchNode) CreateLabelsArray() []models.GitHubPullRequestLabelNode {
	labels := make([]models.GitHubPullRequestLabelNode, len(node.Labels.Edges))
	for i, label := range node.Labels.Edges {
		labels[i] = label.Node
	}
	return labels
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
	Node models.GitHubPullRequestLabelNode `json:"node"`
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
