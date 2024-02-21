// Contains functions to import package mask entries into the database
//
// Example
//
// ## # Dev E. Loper <developer@gentoo.org> (2019-07-01)
// ## # Masking  these versions until we can get the
// ## # v4l stuff to work properly again
// ## =media-video/mplayer-0.90_pre5
// ## =media-video/mplayer-0.90_pre5-r1
//

package repository

import (
	"regexp"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-pg/pg/v10"
)

// isMask checks whether the path
// points to a package.mask file
func isMask(path string) bool {
	return path == "profiles/package.mask"
}

// UpdateMask updates all entries in
// the Mask table in the database
func UpdateMask(path string) {

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

	if status != "D" && isMask(changedFile) {

		logger.Info.Println("Updating Masks")

		// delete all existing masks before parsing the file again
		// in future we might implement a incremental version here
		database.TruncateTable[models.Mask]("versions")

		for _, packageMask := range getMasks(changedFile) {
			parsePackageMask(packageMask)
		}
	}
}

var versionNumber = regexp.MustCompile(`-[0-9]`)

// versionSpecifierToPackageAtom returns the package atom from a given version specifier
func versionSpecifierToPackageAtom(versionSpecifier string) string {
	gpackage := strings.ReplaceAll(versionSpecifier, ">", "")
	gpackage = strings.ReplaceAll(gpackage, "<", "")
	gpackage = strings.ReplaceAll(gpackage, "=", "")
	gpackage = strings.ReplaceAll(gpackage, "~", "")

	gpackage = strings.Split(gpackage, ":")[0]
	gpackage = versionNumber.Split(gpackage, 2)[0]

	return gpackage
}

// parseAuthorLine parses the first line in the package.mask file
// and returns the author name, author email and the date
func parseAuthorLine(authorLine string) (string, string, time.Time) {

	if !(strings.Contains(authorLine, "<") && strings.Contains(authorLine, ">")) {
		logger.Error.Println("Error while parsing the author line in mask entry:", authorLine)
		return "", "", time.Now()
	}

	author := strings.TrimSpace(strings.Split(authorLine, "<")[0])
	author = strings.ReplaceAll(author, "#", "")
	authorEmail := strings.TrimSpace(strings.Split(strings.Split(authorLine, "<")[1], ">")[0])
	date := strings.TrimSpace(strings.Split(authorLine, ">")[1])
	date = strings.ReplaceAll(date, "(", "")
	date = strings.ReplaceAll(date, ")", "")
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		logger.Error.Println("Error while parsing package mask date: " + date)
		logger.Error.Println(err)
	}
	return author, authorEmail, parsedDate
}

// based on regex from GLEP-0084
var bugListMatcher = regexp.MustCompile(`[Bb]ugs? +#\d+(,? +#\d+)*`)
var bugReplacer = regexp.MustCompile(`#(\d+)`)

// parse the package.mask entries and
// update the Mask table in the database
func parsePackageMask(packageMask string) {
	packageMaskLines := strings.Split(packageMask, "\n")
	if len(packageMaskLines) >= 3 {
		packageMaskLine, packageMaskLines := packageMaskLines[0], packageMaskLines[1:]
		author, authorEmail, date := parseAuthorLine(packageMaskLine)

		var reason string
		packageMaskLine, packageMaskLines = packageMaskLines[0], packageMaskLines[1:]
		for strings.HasPrefix(packageMaskLine, "#") {
			if packageMaskLine == "#" {
				reason += "<br />"
			} else {
				reason = reason + " " + templ.EscapeString(strings.TrimPrefix(packageMaskLine, "# "))
			}
			packageMaskLine, packageMaskLines = packageMaskLines[0], packageMaskLines[1:]
		}

		reason = bugListMatcher.ReplaceAllStringFunc(reason, func(bugList string) string {
			return bugReplacer.ReplaceAllString(bugList, `<a href="https://bugs.gentoo.org/$1">$0</a>`)
		})

		packageMaskLines = append(packageMaskLines, packageMaskLine)

		for _, version := range packageMaskLines {
			mask := &models.Mask{
				Author:      strings.TrimSpace(author),
				AuthorEmail: strings.TrimSpace(authorEmail),
				Date:        date,
				Reason:      strings.TrimSpace(reason),
				Versions:    version,
			}

			_, err := database.DBCon.Model(mask).OnConflict("(versions) DO UPDATE").Insert()
			if err != nil {
				logger.Error.Println("Error while inserting/updating package mask entry", err)
			}
		}
	}

}

// get all mask entries from the package.mask file
func getMasks(path string) []string {
	var masks []string
	lines, err := utils.ReadLines(config.PortDir() + "/" + path)

	if err != nil {
		logger.Error.Println("Could not read Masks file. Abort masks import", err)
		return masks
	}

	line, lines := lines[0], lines[1:]
	for !strings.Contains(line, "#--- END OF EXAMPLES ---") {
		line, lines = lines[0], lines[1:]
	}
	lines = lines[1:]

	return strings.Split(strings.Join(lines, "\n"), "\n\n")
}

// Calculate all versions that are currently
// masked and update the MaskToVersion Table
func CalculateMaskedVersions() {
	// clean up all masked versions before recalculating them
	database.TruncateTable[models.MaskToVersion]("id")

	var masks []*models.Mask
	err := database.DBCon.Model(&masks).Select()
	if err != nil && err != pg.ErrNoRows {
		logger.Error.Println("Failed to retrieve package masks. Aborting update", err)
		return
	}

	for _, mask := range masks {
		versionSpecifier := mask.Versions
		packageAtom := versionSpecifierToPackageAtom(versionSpecifier)
		versions := utils.CalculateAffectedVersions(versionSpecifier, packageAtom)
		maskVersions(versionSpecifier, versions)
	}
}

// maskVersions updates the MaskToVersion table using the given versions
func maskVersions(versionSpecifier string, versions []*models.Version) {
	for _, version := range versions {
		maskToVersion := &models.MaskToVersion{
			Id:           versionSpecifier + "-" + version.Id,
			MaskVersions: versionSpecifier,
			VersionId:    version.Id,
		}

		_, err := database.DBCon.Model(maskToVersion).OnConflict("(id) DO UPDATE").Insert()

		if err != nil {
			logger.Error.Println("Error while inserting mask to version entry", err)
		}
	}
}
