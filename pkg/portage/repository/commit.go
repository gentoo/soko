// Contains functions to parse, import and process commits

package repository

import (
	"log/slog"
	"os/exec"
	"slices"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"strings"
	"time"
)

var (
	// arrays for collecting date for batch dump into database
	commits         []*models.Commit
	packages        []*models.Package
	keywordChanges  = map[string]*models.KeywordChange{}
	packagesCommit  []*models.CommitToPackage
	versionsCommits []*models.CommitToVersion
)

// UpdateCommits incrementally imports all new commits. New commits are
// determined by retrieving the last commit in the database (if present)
// and parsing all following commits. In case no last commit is present
// a full import starting with the first commit in the tree is done.
func UpdateCommits() string {
	slog.Info("Start updating commits")

	latestCommit, precedingCommitsOffset := utils.GetLatestCommitAndPreceding()

	for precedingCommits, rawCommit := range utils.GetCommits(latestCommit, "HEAD") {
		latestCommit = processCommit(precedingCommits, precedingCommitsOffset, rawCommit)

		if len(commits) > 10000 {
			dumpToDatabase()
		}
	}
	dumpToDatabase()
	slog.Info("Finished updating commits")

	return latestCommit
}

// processCommit parses a single commit log output and updates it into the database
func processCommit(PrecedingCommits, PrecedingCommitsOffset int, rawCommit string) string {
	commitLines := strings.Split(rawCommit, "\n")

	if len(commitLines) < 8 {
		return ""
	}

	logProgress(PrecedingCommits)

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

	changedFiles := processChangedFiles(PrecedingCommits, PrecedingCommitsOffset, commitLines, id)

	commits = append(commits, &models.Commit{
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
	})
	return id
}

// processChangedFiles parses files that have changed in the commit and links the
// commit to packages and package versions
func processChangedFiles(PrecedingCommits, PrecedingCommitsOffset int, commitLines []string, id string) *models.ChangedFiles {
	var addedFiles, modifiedFiles, deletedFiles []*models.ChangedFile

	for _, commitLine := range commitLines {
		line := strings.Split(commitLine, "\t")
		if len(line) < 2 {
			continue
		}

		status := strings.TrimSpace(line[0])
		path := strings.TrimSpace(line[1])

		if strings.HasPrefix(status, "M") {
			modifiedFiles = append(modifiedFiles, &models.ChangedFile{Path: path, ChangeType: "M"})
			createKeywordChange(id, path, commitLine)
		} else if strings.HasPrefix(commitLine, "D") {
			deletedFiles = append(deletedFiles, &models.ChangedFile{Path: path, ChangeType: "D"})
		} else if strings.HasPrefix(commitLine, "A") {
			addedFiles = append(addedFiles, &models.ChangedFile{Path: path, ChangeType: "A"})
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

// logProgress logs the progress of a loop
func logProgress(counter int) {
	if counter%1000 == 0 {
		slog.Info("Processed commits", slog.Int("commits", counter))
	} else if counter == 1 {
		// The initial commit is *huge* that's why we log it as well
		slog.Info("Processed first commit.")
	}
}

func linkCommitToPackage(commitLine, path, id string) {
	if (len(strings.Split(commitLine, "/")) >= 3) &&
		(strings.HasPrefix(commitLine, "M") ||
			strings.HasPrefix(commitLine, "D") ||
			strings.HasPrefix(commitLine, "A")) {

		pathParts := strings.Split(strings.TrimSuffix(path, ".ebuild"), "/")

		packageAtom := pathParts[0] + "/" + strings.Split(commitLine, "/")[1]
		packagesCommit = append(packagesCommit, &models.CommitToPackage{
			Id:          id + "-" + packageAtom,
			CommitId:    id,
			PackageAtom: packageAtom,
		})
	}
}

func linkCommitToVersion(commitLine, path, id string) {
	if (strings.HasPrefix(commitLine, "M") ||
		strings.HasPrefix(commitLine, "D") ||
		strings.HasPrefix(commitLine, "A")) &&
		len(strings.Split(strings.TrimSuffix(path, ".ebuild"), "/")) == 3 &&
		strings.HasSuffix(strings.TrimSpace(strings.Split(commitLine, "\t")[1]), ".ebuild") {

		pathParts := strings.Split(strings.TrimSuffix(path, ".ebuild"), "/")

		versionId := pathParts[0] + "/" + pathParts[2]
		versionsCommits = append(versionsCommits, &models.CommitToVersion{
			Id:        id + "-" + versionId,
			CommitId:  id,
			VersionId: versionId,
		})
	}
}

func createKeywordChange(id, path, commitLine string) {
	if !strings.HasSuffix(path, ".ebuild") || !(strings.Count(commitLine, "/") >= 2) {
		return
	}

	raw_lines, err := utils.Exec(config.PortDir(), "git", "show", id, "--", path)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); !ok || exitError.ExitCode() != 1 {
			slog.Error("Failed running git show", slog.String("id", id), slog.String("path", path), slog.Any("err", err))
			return
		}
	}

	var keywords_old, keywords_new []string

	for _, line := range raw_lines {
		if strings.HasPrefix(line, "-KEYWORDS=") {
			keywords_old = strings.Split(strings.ReplaceAll(strings.TrimPrefix(line, "-KEYWORDS="), "\"", ""), " ")
		} else if strings.HasPrefix(line, "+KEYWORDS") {
			keywords_new = strings.Split(strings.ReplaceAll(strings.TrimPrefix(line, "+KEYWORDS="), "\"", ""), " ")
		}
	}

	var added_keywords, stabilized_keywords []string

	if keywords_old != nil && keywords_new != nil {
		for _, keyword := range keywords_new {
			if !slices.Contains(keywords_old, keyword) {
				added_keywords = append(added_keywords, keyword)
			}

			if !strings.HasPrefix(keyword, "~") && slices.Contains(keywords_old, "~"+keyword) {
				stabilized_keywords = append(stabilized_keywords, keyword)
			}
		}

		pathParts := strings.Split(strings.TrimSuffix(path, ".ebuild"), "/")

		keywordChangeId := id + "-" + strings.TrimSpace(strings.Split(commitLine, "\t")[1])
		keywordChanges[keywordChangeId] = &models.KeywordChange{
			Id:         keywordChangeId,
			CommitId:   id,
			VersionId:  pathParts[0] + "/" + pathParts[2],
			PackageId:  pathParts[0] + "/" + strings.Split(commitLine, "/")[1],
			Added:      added_keywords,
			Stabilized: stabilized_keywords,
			All:        keywords_new,
		}
	}
}

func createAddedKeywords(id string, path string, commitLine string) {
	if strings.HasSuffix(strings.TrimSpace(strings.Split(commitLine, "\t")[1]), ".ebuild") &&
		(strings.Count(commitLine, "/") >= 2) {

		raw_lines, err := utils.Exec(config.PortDir(), "git", "show", id, "--", path)
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); !ok || exitError.ExitCode() != 1 {
				slog.Error("Failed running git show", slog.String("id", id), slog.String("path", path), slog.Any("err", err))
				return
			}
		}

		for _, line := range raw_lines {
			if strings.HasPrefix(line, "+KEYWORDS=") {
				pathParts := strings.Split(strings.TrimSuffix(path, ".ebuild"), "/")
				keywords := strings.Split(strings.ReplaceAll(strings.TrimPrefix(line, "+KEYWORDS="), "\"", ""), " ")

				keywordChangeId := id + "-" + strings.TrimSpace(strings.Split(commitLine, "\t")[1])
				keywordChanges[keywordChangeId] = &models.KeywordChange{
					Id:        keywordChangeId,
					CommitId:  id,
					VersionId: pathParts[0] + "/" + pathParts[2],
					PackageId: pathParts[0] + "/" + strings.Split(commitLine, "/")[1],
					Added:     keywords,
					All:       keywords,
				}
			}
		}

	}
}

func updateFirstCommitOfPackage(path string, commitLine string, precedingCommits int) {
	// Added Package
	if strings.HasSuffix(path, "metadata.xml") && strings.Count(commitLine, "/") == 2 {
		atom := strings.Split(path, "/")[0] + "/" + strings.Split(path, "/")[1]
		packages = append(packages, &models.Package{
			Atom:             atom,
			PrecedingCommits: precedingCommits,
		})
	}
}

func dumpToDatabase() {
	slog.Info("Writing to database",
		slog.Int("KeywordChange", len(keywordChanges)),
		slog.Int("Package", len(packages)),
		slog.Int("CommitToPackage", len(packagesCommit)),
		slog.Int("CommitToVersion", len(versionsCommits)),
		slog.Int("Commit", len(commits)))

	if len(keywordChanges) > 0 {
		rows := make([]*models.KeywordChange, 0, len(keywordChanges))
		for _, keywordChange := range keywordChanges {
			rows = append(rows, keywordChange)
		}
		_, err := database.DBCon.Model(&rows).OnConflict("(id) DO UPDATE").Insert()
		if err != nil {
			slog.Error("Failed inserting KeywordChange", slog.Any("err", err))
		}
		clear(keywordChanges)
	}

	if len(packages) > 0 {
		_, err := database.DBCon.Model(&packages).Column("preceding_commits").Update()
		if err != nil {
			slog.Error("Failed inserting Package", slog.Any("err", err))
		}
		clear(packages)
	}

	if len(packagesCommit) > 0 {
		_, err := database.DBCon.Model(&packagesCommit).OnConflict("(id) DO NOTHING").Insert()
		if err != nil {
			slog.Error("Failed inserting CommitToPackage", slog.Any("err", err))
		}
		clear(packagesCommit)
	}

	if len(versionsCommits) > 0 {
		_, err := database.DBCon.Model(&versionsCommits).OnConflict("(id) DO NOTHING").Insert()
		if err != nil {
			slog.Error("Failed inserting CommitToVersion", slog.Any("err", err))
		}
		clear(versionsCommits)
	}

	if len(commits) > 0 {
		_, err := database.DBCon.Model(&commits).OnConflict("(id) DO UPDATE").Insert()
		if err != nil {
			slog.Error("Failed inserting Commit", slog.Any("err", err))
		}
		clear(commits)
	}
}
