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
			reason = reason + " " + strings.Replace(packageMaskLine, "# ", "", 1)
			packageMaskLine, packageMaskLines = packageMaskLines[0], packageMaskLines[1:]
		}

		packageMaskLines = append(packageMaskLines, packageMaskLine)

		for _, version := range packageMaskLines {
			mask := &models.Mask{
				Author:      author,
				AuthorEmail: authorEmail,
				Date:        date,
				Reason:      reason,
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
		var versions []*models.Version

		if strings.HasPrefix(versionSpecifier, "=") {
			versions = exactVersion(versionSpecifier, packageAtom)
		} else if strings.HasPrefix(versionSpecifier, "<=") {
			versions = comparedVersions("<=", versionSpecifier, packageAtom)
		} else if strings.HasPrefix(versionSpecifier, "<") {
			versions = comparedVersions("<", versionSpecifier, packageAtom)
		} else if strings.HasPrefix(versionSpecifier, ">=") {
			versions = comparedVersions(">=", versionSpecifier, packageAtom)
		} else if strings.HasPrefix(versionSpecifier, ">") {
			versions = comparedVersions(">", versionSpecifier, packageAtom)
		} else if strings.HasPrefix(versionSpecifier, "~") {
			versions = allRevisions(versionSpecifier, packageAtom)
		} else if strings.Contains(versionSpecifier, ":") {
			versions = versionsWithSlot(versionSpecifier, packageAtom)
		} else {
			versions = allVersions(versionSpecifier, packageAtom)
		}

		maskVersions(versionSpecifier, versions)
	}
}

// extract slot and subslot name from versionSpecifier
func slotAndSubslot(versionSpecifier string) (string, []string) {
	version, fullslot, found := strings.Cut(versionSpecifier, ":")
	if found {
		return version, strings.SplitN(fullslot, "/", 2)
	} else {
		return version, nil
	}
}

// comparedVersions computes and returns all versions that are >=, >, <= or < than then given version
func comparedVersions(operator string, versionSpecifier string, packageAtom string) []*models.Version {
	var results, versions []*models.Version
	versionSpecifier = strings.ReplaceAll(versionSpecifier, operator, "")
	versionSpecifier = strings.ReplaceAll(versionSpecifier, packageAtom+"-", "")
	versionSpecifier, slots := slotAndSubslot(versionSpecifier)

	q := database.DBCon.Model(&versions).
		Where("atom = ?", packageAtom)
	if len(slots) >= 1 {
		q = q.Where("slot = ?", slots[0])
	}
	if len(slots) == 2 {
		q = q.Where("subslot = ?", slots[1])
	}
	q.Select()

	for _, v := range versions {
		givenVersion := models.Version{Version: versionSpecifier}
		if operator == ">" {
			if v.GreaterThan(givenVersion) {
				results = append(results, v)
			}
		} else if operator == ">=" {
			if v.GreaterThan(givenVersion) || v.EqualTo(givenVersion) {
				results = append(results, v)
			}
		} else if operator == "<" {
			if v.SmallerThan(givenVersion) {
				results = append(results, v)
			}
		} else if operator == "<=" {
			if v.SmallerThan(givenVersion) || v.EqualTo(givenVersion) {
				results = append(results, v)
			}
		}
	}
	return results
}

var revision = regexp.MustCompile(`-r[0-9]*$`)

// allRevisions returns all revisions of the given version
func allRevisions(versionSpecifier string, packageAtom string) []*models.Version {
	var versions []*models.Version
	versionSpecifier, slots := slotAndSubslot(versionSpecifier)
	versionWithoutRevision := revision.Split(versionSpecifier, 1)[0]
	versionWithoutRevision = strings.ReplaceAll(versionWithoutRevision, "~", "")

	q := database.DBCon.Model(&versions).
		Where("id LIKE ?", versionWithoutRevision+"%")
	if len(slots) >= 1 {
		q = q.Where("slot = ?", slots[0])
	}
	if len(slots) == 2 {
		q = q.Where("subslot = ?", slots[1])
	}
	q.Select()

	return versions
}

// exactVersion returns the exact version specified in the versionSpecifier
func exactVersion(versionSpecifier string, packageAtom string) []*models.Version {
	var versions []*models.Version
	versionSpecifier, slots := slotAndSubslot(versionSpecifier)

	q := database.DBCon.Model(&versions).
		Where("id = ?", strings.Replace(versionSpecifier, "=", "", 1))
	if len(slots) >= 1 {
		q = q.Where("slot = ?", slots[0])
	}
	if len(slots) == 2 {
		q = q.Where("subslot = ?", slots[1])
	}
	q.Select()

	return versions
}

// versionsWithSlot returns all versions with the given slot
func versionsWithSlot(versionSpecifier string, packageAtom string) []*models.Version {
	var versions []*models.Version
	_, slots := slotAndSubslot(versionSpecifier)

	q := database.DBCon.Model(&versions).
		Where("atom = ?", packageAtom).
		Where("slot = ?", slots[0])
	if len(slots) == 2 {
		q = q.Where("subslot = ?", slots[1])
	}
	q.Select()

	return versions
}

// allVersions returns all versions of the given package
func allVersions(versionSpecifier string, packageAtom string) []*models.Version {
	var versions []*models.Version
	database.DBCon.Model(&versions).
		Where("atom = ?", packageAtom).
		Select()
	return versions
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
