// SPDX-License-Identifier: GPL-2.0-only

// Contains functions to import package deprecated entries into the database
//
// Example
//
// ## # Dev E. Loper <developer@gentoo.org> (2019-07-01)
// ## # Deprecated upstream, see HOMEPAGE
// ## dev-perl/Mail-Sender
//

package repository

import (
	"log/slog"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-pg/pg/v10"
)

// isPackagesDeprecated checks whether the path
// points to a package.mask file
func isPackagesDeprecated(path string) bool {
	return path == "profiles/package.deprecated"
}

// UpdatePackagesDeprecated updates all entries in
// the Deprecated table in the database
func UpdatePackagesDeprecated(path string) {

	splittedLine := strings.Split(path, "\t")

	var status, changedFile string
	switch len(splittedLine) {
	case 2:
		status = splittedLine[0]
		changedFile = splittedLine[1]
	case 1:
		// This happens in case of a full update
		status = "A"
		changedFile = splittedLine[0]
	default:
		// should not happen
		return
	}

	if status != "D" && isPackagesDeprecated(changedFile) {
		slog.Info("Updating package.deprecated")

		// delete all existing entries before parsing the file again
		// in future we might implement a incremental version here
		database.TruncateTable((*models.DeprecatedPackage)(nil))

		for _, entry := range getDeprecatedPackages(changedFile) {
			parsePackagesDeprecated(entry)
		}
	}
}

// parse the package.mask entries and
// update the DeprecatedPackage table in the database
func parsePackagesDeprecated(entry string) {
	packageLines := strings.Split(entry, "\n")
	if len(packageLines) >= 3 {
		packageLine, packageLines := packageLines[0], packageLines[1:]
		author, authorEmail, date := parseAuthorLine(packageLine)

		var reason string
		packageLine, packageLines = packageLines[0], packageLines[1:]
		for strings.HasPrefix(packageLine, "#") {
			if packageLine == "#" {
				reason += "<br />"
			} else {
				reason = reason + " " + templ.EscapeString(strings.TrimPrefix(packageLine, "# "))
			}
			if len(packageLines) == 0 {
				break
			}
			packageLine, packageLines = packageLines[0], packageLines[1:]
		}

		reason = bugListMatcher.ReplaceAllStringFunc(reason, func(bugList string) string {
			return bugReplacer.ReplaceAllString(bugList, `<a href="https://bugs.gentoo.org/$1">$0</a>`)
		})

		packageLines = append(packageLines, packageLine)

		for _, version := range packageLines {
			entry := &models.DeprecatedPackage{
				Author:      strings.TrimSpace(author),
				AuthorEmail: strings.TrimSpace(authorEmail),
				Date:        date,
				Reason:      strings.TrimSpace(reason),
				Versions:    version,
			}

			_, err := database.DBCon.Model(entry).OnConflict("(versions) DO UPDATE").Insert()
			if err != nil {
				slog.Error("Failed inserting/updating package deprecated entry", slog.Any("err", err))
			}
		}
	}

}

// get all entries from the package.deprecated file
func getDeprecatedPackages(path string) []string {
	var deprecates []string
	lines, err := utils.ReadLines(config.PortDir() + "/" + path)
	if err != nil {
		slog.Error("Could not read package.deprecated file. Abort deprecated import", slog.Any("err", err))
		return deprecates
	}

	line, lines := lines[0], lines[1:]
	for !strings.Contains(line, "#--- END OF EXAMPLES ---") {
		line, lines = lines[0], lines[1:]
	}
	lines = lines[1:]

	return strings.Split(strings.Join(lines, "\n"), "\n\n")
}

// Calculate all versions that are currently
// deprecated and update the DeprecatedToVersion Table
func CalculateDeprecatedToVersion() {
	database.TruncateTable((*models.DeprecatedToVersion)(nil))

	var deprecates []*models.DeprecatedPackage
	err := database.DBCon.Model(&deprecates).Select()
	if err != nil && err != pg.ErrNoRows {
		slog.Error("Failed to retrieve package masks. Aborting update", slog.Any("err", err))
		return
	}

	for _, deprecate := range deprecates {
		versionSpecifier := deprecate.Versions
		packageAtom := versionSpecifierToPackageAtom(versionSpecifier)
		versions := utils.CalculateAffectedVersions(versionSpecifier, packageAtom)

		for _, version := range versions {
			depToVersion := &models.DeprecatedToVersion{
				Id:                 versionSpecifier + "-" + version.Id,
				DeprecatedVersions: versionSpecifier,
				VersionId:          version.Id,
			}

			_, err := database.DBCon.Model(depToVersion).OnConflict("(id) DO UPDATE").Insert()
			if err != nil {
				slog.Error("Failed inserting/updating deprecated to version entry", slog.Any("err", err))
			}
		}
	}
}
