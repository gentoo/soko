// Contains functions to parse, import and process commits

package repository

import (
	"os/exec"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"strconv"
	"strings"
	"time"
)

// UpdateCommits incrementally imports all new commits. New commits are
// determined by retrieving the last commit in the database (if present)
// and parsing all following commits. In case no last commit is present
// a full import starting with the first commit in the tree is done.
func UpdateCommits() string {
	logger.Info.Println("Start updating commits")

	latestCommit, PrecedingCommitsOffset := utils.GetLatestCommitAndPreceeding()

	for PrecedingCommits, rawCommit := range utils.GetCommits(latestCommit, "HEAD") {
		latestCommit = processCommit(PrecedingCommits, PrecedingCommitsOffset, rawCommit)
	}
	logger.Info.Println("Finished updating commits")

	return latestCommit
}

// processCommit parses a single commit log output and updates it into the database
func processCommit(PrecedingCommits int, PrecedingCommitsOffset int, rawCommit string) string {

	commitLines := strings.Split(rawCommit, "\n")

	if len(commitLines) < 8 {
		return ""
	}

	logProgess(PrecedingCommits)

	id := strings.TrimSpace(strings.ReplaceAll(commitLines[0], "commit ", ""))
	authorName := strings.TrimSpace(strings.Split(strings.ReplaceAll(commitLines[1], "Author: ", ""), "<")[0])
	authorEmail := strings.TrimSpace(strings.ReplaceAll(strings.Split(strings.ReplaceAll(commitLines[1], "Author: ", ""), "<")[1], ">", ""))
	rawAuthorDate := strings.TrimSpace(strings.ReplaceAll(commitLines[2], "AuthorDate: ", ""))
	parsedAuthorDate, _ := time.Parse(time.RFC3339, strings.ReplaceAll(rawAuthorDate[:23]+":"+rawAuthorDate[23:], "'", ""))
	authorDate := parsedAuthorDate
	committerName := strings.TrimSpace(strings.Split(strings.ReplaceAll(commitLines[3], "Commit: ", ""), "<")[0])
	committerEmail := strings.TrimSpace(strings.ReplaceAll(strings.Split(strings.ReplaceAll(commitLines[3], "Commit: ", ""), "<")[1], ">", ""))
	rawCommitterDate := strings.TrimSpace(strings.ReplaceAll(commitLines[4], "CommitDate: ", ""))
	parsedCommitterDate, _ := time.Parse(time.RFC3339, strings.ReplaceAll(rawCommitterDate[:23]+":"+rawCommitterDate[23:], "'", ""))
	committerDate := parsedCommitterDate
	message := strings.TrimSpace(commitLines[6])

	commitLines = commitLines[7:]

	if authorEmail == "repomirrorci@gentoo.org" || authorEmail == "repo-qa-checks@gentoo.org" {
		return id
	}

	changedFiles := processChangedFiles(PrecedingCommits, PrecedingCommitsOffset, commitLines, id)

	commit := &models.Commit{
		Id:               id,
		PrecedingCommits: PrecedingCommitsOffset + PrecedingCommits + 1,
		AuthorName:       authorName,
		AuthorEmail:      authorEmail,
		AuthorDate:       authorDate,
		CommitterName:    committerName,
		CommitterEmail:   committerEmail,
		CommitterDate:    committerDate,
		Message:          message,
		ChangedFiles:     changedFiles,
	}

	_, err := database.DBCon.Model(commit).OnConflict("(id) DO UPDATE").Insert()

	if err != nil {
		logger.Error.Println("Error during updating commit: " + id)
		logger.Error.Println(err)
	}
	return id
}

// processChangedFiles parses files that have changed in the commit and links the
// commit to packages and package versions
func processChangedFiles(PrecedingCommits int, PrecedingCommitsOffset int, commitLines []string, id string) *models.ChangedFiles {
	var addedFiles []*models.ChangedFile
	var modifiedFiles []*models.ChangedFile
	var deletedFiles []*models.ChangedFile

	for _, commitLine := range commitLines {

		line := strings.Split(commitLine, "\t")
		if len(line) < 2 {
			continue
		}

		status := strings.TrimSpace(line[0])
		path := strings.TrimSpace(line[1])

		if strings.HasPrefix(status, "M") {

			modifiedFiles = addChangedFile(modifiedFiles, path, "M")
			createKeywordChange(id, path, commitLine)

		} else if strings.HasPrefix(commitLine, "D") {

			deletedFiles = addChangedFile(deletedFiles, path, "D")

		} else if strings.HasPrefix(commitLine, "A") {

			addedFiles = addChangedFile(addedFiles, path, "A")
			updateFirstCommitOfPackage(path, commitLine, PrecedingCommitsOffset+PrecedingCommits+1)
			createAddedKeywords(id, path, commitLine)

		}

		linkCommitToPackage(commitLine, path, id)
		linkCommitToVersion(commitLine, path, id)

	}

	return &models.ChangedFiles{
		Added:    addedFiles,
		Modified: modifiedFiles,
		Deleted:  deletedFiles,
	}
}

// logProgess logs the progress of a loop
func logProgess(counter int) {
	if counter%1000 == 0 {
		logger.Info.Println("Processed commits: " + strconv.Itoa(counter))
	} else if counter == 1 {
		// The initial commit is *huge* that's why we log it as well
		logger.Info.Println("Processed first commit.")
	}
}

func linkCommitToPackage(commitLine string, path string, id string) {
	var commitToPackage *models.CommitToPackage
	if (len(strings.Split(commitLine, "/")) >= 3) &&
		(strings.HasPrefix(commitLine, "M") ||
			strings.HasPrefix(commitLine, "D") ||
			strings.HasPrefix(commitLine, "A")) {

		pathParts := strings.Split(strings.ReplaceAll(path, ".ebuild", ""), "/")

		commitToPackageId := id + "-" + pathParts[0] + "/" + strings.Split(commitLine, "/")[1]
		commitToPackage = &models.CommitToPackage{
			Id:          commitToPackageId,
			CommitId:    id,
			PackageAtom: pathParts[0] + "/" + strings.Split(commitLine, "/")[1],
		}

		_, err := database.DBCon.Model(commitToPackage).OnConflict("(id) DO NOTHING").Insert()

		if err != nil {
			logger.Error.Println("Error during updating CommitToPackage: " + commitToPackageId)
			logger.Error.Println(err)
		}

	}
}

func linkCommitToVersion(commitLine string, path string, id string) {
	var commitToVersion *models.CommitToVersion
	if (strings.HasPrefix(commitLine, "M") ||
		strings.HasPrefix(commitLine, "D") ||
		strings.HasPrefix(commitLine, "A")) &&
		len(strings.Split(strings.ReplaceAll(path, ".ebuild", ""), "/")) == 3 &&
		strings.HasSuffix(strings.TrimSpace(strings.Split(commitLine, "\t")[1]), ".ebuild") {

		pathParts := strings.Split(strings.ReplaceAll(path, ".ebuild", ""), "/")

		commitToVersionId := id + "-" + pathParts[0] + "/" + pathParts[2]
		commitToVersion = &models.CommitToVersion{
			Id:        commitToVersionId,
			CommitId:  id,
			VersionId: pathParts[0] + "/" + pathParts[2],
		}

		_, err := database.DBCon.Model(commitToVersion).OnConflict("(id) DO NOTHING").Insert()

		if err != nil {
			logger.Error.Println("Error during updating CommitToVersion: " + commitToVersionId)
			logger.Error.Println(err)
		}

	}
}

func createKeywordChange(id string, path string, commitLine string) {

	if !strings.HasSuffix(path, ".ebuild") || !(len(strings.Split(commitLine, "/")) >= 3) {
		return
	}

	var change *models.KeywordChange

	raw_lines, err := utils.Exec(config.PortDir(), "git", "show", id, "--", path)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); !ok || exitError.ExitCode() != 1 {
			logger.Error.Println("Problem parsing file")
			return
		}
	}

	var keywords_old []string
	var keywords_new []string

	for _, line := range raw_lines {
		if strings.HasPrefix(line, "-KEYWORDS=") {
			keywords_old = strings.Split(strings.ReplaceAll(strings.ReplaceAll(line, "-KEYWORDS=", ""), "\"", ""), " ")

		} else if strings.HasPrefix(line, "+KEYWORDS") {
			keywords_new = strings.Split(strings.ReplaceAll(strings.ReplaceAll(line, "+KEYWORDS=", ""), "\"", ""), " ")
		}
	}

	var added_keywords []string
	var stabilized_keywords []string

	if keywords_old != nil && keywords_new != nil {

		for _, keyword := range keywords_new {
			if !utils.Contains(keywords_old, keyword) {
				added_keywords = append(added_keywords, keyword)
			}

			if !strings.HasPrefix(keyword, "~") && utils.Contains(keywords_old, ("~"+keyword)) {
				stabilized_keywords = append(stabilized_keywords, keyword)
			}
		}

		pathParts := strings.Split(strings.ReplaceAll(path, ".ebuild", ""), "/")

		keywordChangeId := id + "-" + strings.TrimSpace(strings.Split(commitLine, "\t")[1])
		change = &models.KeywordChange{
			Id:         keywordChangeId,
			CommitId:   id,
			VersionId:  pathParts[0] + "/" + pathParts[2],
			PackageId:  pathParts[0] + "/" + strings.Split(commitLine, "/")[1],
			Added:      added_keywords,
			Stabilized: stabilized_keywords,
			All:        keywords_new,
		}

		_, err := database.DBCon.Model(change).OnConflict("(id) DO UPDATE").Insert()

		if err != nil {
			logger.Error.Println("Error updating Keyword change: " + keywordChangeId)
			logger.Error.Println(err)
		}

	}
}

func createAddedKeywords(id string, path string, commitLine string) {
	var change *models.KeywordChange
	if strings.HasSuffix(strings.TrimSpace(strings.Split(commitLine, "\t")[1]), ".ebuild") &&
		(len(strings.Split(commitLine, "/")) >= 3) {

		raw_lines, err := utils.Exec(config.PortDir(), "git", "show", id, "--", path)
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); !ok || exitError.ExitCode() != 1 {
				logger.Error.Println("Problem parsing file")
				logger.Error.Println(exitError)
				return
			}
		}

		for _, line := range raw_lines {
			if strings.HasPrefix(line, "+KEYWORDS=") {

				pathParts := strings.Split(strings.ReplaceAll(path, ".ebuild", ""), "/")
				keywords := strings.Split(strings.ReplaceAll(strings.ReplaceAll(line, "+KEYWORDS=", ""), "\"", ""), " ")

				keywordChangeId := id + "-" + strings.TrimSpace(strings.Split(commitLine, "\t")[1])
				change = &models.KeywordChange{
					Id:        keywordChangeId,
					CommitId:  id,
					VersionId: pathParts[0] + "/" + pathParts[2],
					PackageId: pathParts[0] + "/" + strings.Split(commitLine, "/")[1],
					Added:     keywords,
					All:       keywords,
				}

				_, err := database.DBCon.Model(change).OnConflict("(id) DO UPDATE").Insert()

				if err != nil {
					logger.Error.Println("Error updating Keyword change: " + keywordChangeId)
					logger.Error.Println(err)
				}

			}
		}

	}
}

func updateFirstCommitOfPackage(path string, commitLine string, precedingCommits int) {
	// Added Package
	if strings.HasSuffix(path, "metadata.xml") && len(strings.Split(path, "/")) == 3 {

		atom := strings.Split(path, "/")[0] + "/" + strings.Split(path, "/")[1]
		addedpackage := &models.Package{
			Atom:             atom,
			PrecedingCommits: precedingCommits,
		}

		_, err := database.DBCon.Model(addedpackage).Column("preceding_commits").WherePK().Update()
		if err != nil {
			logger.Error.Println("Error updating precedingCommits (" + strconv.Itoa(precedingCommits) + ") of package: " + atom)
			logger.Error.Println(err)
		}

	}
}

func addChangedFile(changedFiles []*models.ChangedFile, path string, status string) []*models.ChangedFile {
	return append(changedFiles, &models.ChangedFile{
		Path:       path,
		ChangeType: status,
	})
}
