// SPDX-License-Identifier: GPL-2.0-only
// Contains utility functions to parse the output of git commands

package repository

import (
	"os"
	"soko/pkg/config"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func UpdateRepo() error {

	r, err := git.PlainOpen(config.SelfCheckPortDir())

	if err != nil {
		r, _ = git.PlainClone(config.SelfCheckPortDir(), false, &git.CloneOptions{
			URL:      "https://github.com/gentoo-mirror/gentoo",
			Depth:    5,
			Progress: os.Stdout,
		})
	}

	w, _ := r.Worktree()
	err = w.Pull(&git.PullOptions{RemoteName: "origin", ReferenceName: "stable"})

	return err
}

func AllFiles() []string {
	var allFiles []string

	revision := "stable"

	r, _ := git.PlainOpen(config.SelfCheckPortDir())

	h, _ := r.ResolveRevision(plumbing.Revision(revision))

	commit, _ := r.CommitObject(*h)

	tree, _ := commit.Tree()

	// ... get the files iterator and print the file
	tree.Files().ForEach(func(f *object.File) error {
		//fmt.Printf("100644 blob %s    %s\n", f.Hash, f.Name)
		allFiles = append(allFiles, f.Name)
		return nil
	})

	return allFiles
}
