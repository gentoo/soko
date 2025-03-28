// SPDX-License-Identifier: GPL-2.0-only
package models

import (
	"strconv"
)

type GithubPullRequest struct {
	Id          string `pg:",pk"`
	Closed      bool
	Url         string
	Title       string
	CreatedAt   string
	UpdatedAt   string
	CiState     string
	CiStateLink string
	Labels      []GitHubPullRequestLabelNode
	Comments    int
	Files       []GitHubPullRequestFileNode
	Author      string
}

type PackageToGithubPullRequest struct {
	Id                  string `pg:",pk"`
	PackageAtom         string
	GithubPullRequestId string
}

// -- raw json models

type GitHubPullRequestQueryResult struct {
	Data GitHubPullRequestQueryResultData `json:"data"`
}

func (res *GitHubPullRequestQueryResult) HasNextPage() bool {
	return res.Data.Search.PageInfo.HasNextPage
}

func (res *GitHubPullRequestQueryResult) EndCursor() string {
	return res.Data.Search.PageInfo.EndCursor
}

func (res *GitHubPullRequestQueryResult) AppendPullRequest(pullRequests map[int]*GithubPullRequest) {
	for _, rawObject := range res.Data.Search.Edges {
		pullRequest := rawObject.Node
		var ciState, ciStateLink string
		if nodes := pullRequest.Commits.Nodes; len(nodes) > 0 {
			ciState = nodes[0].Commit.Status.State

			if contexts := nodes[0].Commit.Status.Contexts; len(contexts) > 0 {
				ciStateLink = contexts[0].TargetUrl
			}
		}

		pullRequests[pullRequest.Number] = &GithubPullRequest{
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
		}
	}
}

func (node *GitHubPullRequestSearchNode) CreateLabelsArray() []GitHubPullRequestLabelNode {
	labels := make([]GitHubPullRequestLabelNode, len(node.Labels.Edges))
	for i, label := range node.Labels.Edges {
		labels[i] = label.Node
	}
	return labels
}

func (node *GitHubPullRequestSearchNode) CreateFilesArray() []GitHubPullRequestFileNode {
	labels := make([]GitHubPullRequestFileNode, len(node.Files.Edges))
	for i, label := range node.Files.Edges {
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
