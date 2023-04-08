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
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"strings"

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
		logger.Info.Println("Updating package.deprecated")

		// delete all existing entries before parsing the file again
		// in future we might implement a incremental version here
		database.TruncateTable[models.DeprecatedPackage]("versions")

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
			reason = reason + " " + strings.Replace(packageLine, "# ", "", 1)
			packageLine, packageLines = packageLines[0], packageLines[1:]
		}

		packageLines = append(packageLines, packageLine)

		for _, version := range packageLines {
			entry := &models.DeprecatedPackage{
				Author:      author,
				AuthorEmail: authorEmail,
				Date:        date,
				Reason:      reason,
				Versions:    version,
			}

			_, err := database.DBCon.Model(entry).OnConflict("(versions) DO UPDATE").Insert()
			if err != nil {
				logger.Error.Println("Error while inserting/updating package deprecated entry", err)
			}
		}
	}

}

// get all entries from the package.deprecated file
func getDeprecatedPackages(path string) []string {
	var deprecates []string
	lines, err := utils.ReadLines(config.PortDir() + "/" + path)
	if err != nil {
		logger.Error.Println("Could not read package.deprecated file, aborting import, err:", err)
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
	database.TruncateTable[models.DeprecatedToVersion]("id")

	var deprecates []*models.DeprecatedPackage
	err := database.DBCon.Model(&deprecates).Select()
	if err != nil && err != pg.ErrNoRows {
		logger.Error.Println("Failed to retrieve package masks. Aborting update", err)
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
				logger.Error.Println("Error while inserting mask to version entry", err)
			}
		}
	}
}
