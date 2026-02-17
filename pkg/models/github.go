// SPDX-License-Identifier: GPL-2.0-only
package models

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

type GitHubPullRequestLabelNode struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type GitHubPullRequestFileNode struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}
