// SPDX-License-Identifier: GPL-2.0-only
package codeberg

import (
	"time"
)

type apiUser struct {
	Login string `json:"login"`
}

type apiLabel struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type apiPullRequest struct {
	ID        int64      `json:"id"`
	Number    int64      `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"`
	HTMLURL   string     `json:"html_url"`
	User      apiUser    `json:"user"`
	Labels    []apiLabel `json:"labels"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// In Gitea this is commonly present (PRs are issues underneath).
	Comments int `json:"comments"`

	Head struct {
		Sha string `json:"sha"`
	} `json:"head"`
}

type apiPRFile struct {
	Filename string `json:"filename"`
}

type apiCommitStatus struct {
	Status    string    `json:"status"`     // success, failure, pending, error, warning (varies)
	Context   string    `json:"context"`    // e.g. "gentoo-ci"
	TargetURL string    `json:"target_url"` // details link
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
