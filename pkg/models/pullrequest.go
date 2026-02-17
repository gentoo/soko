// SPDX-License-Identifier: GPL-2.0-only
package models

import "iter"

type PullRequest struct {
	Id          string `pg:",pk"`
	Closed      bool
	Url         string
	Title       string
	CreatedAt   string
	UpdatedAt   string
	CiState     string
	CiStateLink string
	Labels      []PullRequestLabel
	Comments    int
	Author      string
}

type PackageToPullRequest struct {
	Id            string `pg:",pk"`
	PackageAtom   string
	PullRequestId string
}

type PullRequestLabel struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type PullRequestProvider interface {
	ToPullRequest() *PullRequest
	GetFiles() iter.Seq[string]
}
