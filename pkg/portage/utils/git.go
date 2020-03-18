// Contains utility functions to parse the output of git commands

package utils

import (
	"log"
	"os/exec"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
)

// ChangedFiles returns a list of files that have been changed
// between the startCommit and the endCommit. The status of the
// change as well as the path to the file is returned for each file
func ChangedFiles(startCommit string, endCommit string) []string {

	cmd := exec.Command("git", "--no-pager",
		                "diff",
		                "--name-status",
		                startCommit + ".." + endCommit)

	cmd.Dir = config.PortDir()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	return strings.Split(string(out),"\n")
}

// GetCommits returns the log message of all commits after
// the given startCommit and before the given endCommit. The
// log message:
//  - uses '%Y-%m-%dT%H:%M:%S%z' as date format
//  - doesn't include merges
//  - doesn't include renames
//  - includes the status of the changed files
// Furthermore the commits are in reverse order.
func GetCommits(startCommit string, endCommit string) []string {

	cmd := exec.Command("git", "--no-pager",
		                "log",
		                "--name-status",
		                "--no-renames",
		                "--no-merges",
		                "--date=format:'%Y-%m-%dT%H:%M:%S%z'",
		                "--format=fuller",
		                "--reverse",
		                startCommit + ".." + endCommit)

	cmd.Dir = config.PortDir()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	return strings.Split(string(out),"\n\ncommit")
}

// GetLatestCommit retrieves the latest commit in
// the database and returns the hash of the commit
func GetLatestCommit() string {
	latestCommit, _ := GetLatestCommitAndPreceeding()
	return latestCommit
}

// GetLatestCommitAndPreceeding retrieves the latest
// commit in the database. The hash of the latest commit
// as well as the number of preceding commits is returned
func GetLatestCommitAndPreceeding() (string, int) {
	latestCommit := EmptyTree()
	PrecedingCommitsOffset := 0

	var commits []*models.Commit
	err := database.DBCon.Model(&commits).
		Order("preceding_commits DESC").
		Limit(1).
		Select()
	if err == nil && len(commits) == 1 {
		latestCommit = commits[0].Id
		PrecedingCommitsOffset = commits[0].PrecedingCommits
	}

	return latestCommit, PrecedingCommitsOffset
}

// EmptyTree returns the hash of the empty tree
func EmptyTree() string {
	return "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
}
