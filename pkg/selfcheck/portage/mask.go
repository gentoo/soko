// SPDX-License-Identifier: GPL-2.0-only
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
	"log/slog"
	"regexp"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"soko/pkg/selfcheck/storage"
	"strings"
	"time"
)

// isMask checks whether the path
// points to a package.mask file
func isMask(path string) bool {
	return path == "profiles/package.mask"
}

// UpdateMask updates all entries in
// the Mask table in the database
func UpdateMask(path string) {
	if isMask(path) {
		for _, packageMask := range getMasks(path) {
			parsePackageMask(packageMask)
		}
	}
}

// versionSpecifierToPackageAtom returns the package atom from a given version specifier
func versionSpecifierToPackageAtom(versionSpecifier string) string {
	gpackage := strings.ReplaceAll(versionSpecifier, ">", "")
	gpackage = strings.ReplaceAll(gpackage, "<", "")
	gpackage = strings.ReplaceAll(gpackage, "=", "")
	gpackage = strings.ReplaceAll(gpackage, "~", "")

	gpackage = strings.Split(gpackage, ":")[0]

	versionnumber := regexp.MustCompile(`-[0-9]`)
	gpackage = versionnumber.Split(gpackage, 2)[0]

	return gpackage
}

// parseAuthorLine parses the first line in the package.mask file
// and returns the author name, author email and the date
func parseAuthorLine(authorLine string) (string, string, time.Time) {

	if !(strings.Contains(authorLine, "<") && strings.Contains(authorLine, ">")) {
		slog.Error("Error while parsing the author line in mask entry", slog.String("authorLine", authorLine))
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
		slog.Error("Failed parsing package mask date", slog.String("date", date), slog.Any("err", err))
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

		reason := ""
		packageMaskLine, packageMaskLines = packageMaskLines[0], packageMaskLines[1:]
		for strings.HasPrefix(packageMaskLine, "#") {
			reason = reason + " " + strings.Replace(packageMaskLine, "# ", "", 1)
			packageMaskLine, packageMaskLines = packageMaskLines[0], packageMaskLines[1:]
		}

		packageMaskLines = append(packageMaskLines, packageMaskLine)

		for _, version := range packageMaskLines {
			useflag := &models.Mask{
				Author:      author,
				AuthorEmail: authorEmail,
				Date:        date,
				Reason:      reason,
				Versions:    version,
			}

			storage.Masks = append(storage.Masks, useflag)

		}
	}

}

// get all mask entries from the package.mask file
func getMasks(path string) []string {
	var masks []string
	lines, err := utils.ReadLines(path)

	if err != nil {
		slog.Error("Could not read Masks file. Abort masks import", slog.Any("err", err))
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

	for _, mask := range storage.Masks {
		versionSpecifier := mask.Versions
		packageAtom := versionSpecifierToPackageAtom(versionSpecifier)
		var versions []*models.Version

		if strings.HasPrefix(versionSpecifier, "=") {
			versions = exaktVersion(versionSpecifier, packageAtom)
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

		maskVersions(mask, versions)
	}
}

// comparedVersions computes and returns all versions that are >=, >, <= or < than then given version
func comparedVersions(operator string, versionSpecifier string, packageAtom string) []*models.Version {
	var results []*models.Version
	var versions []*models.Version
	versionSpecifier = strings.ReplaceAll(versionSpecifier, operator, "")
	versionSpecifier = strings.ReplaceAll(versionSpecifier, packageAtom+"-", "")
	versionSpecifier = strings.Split(versionSpecifier, ":")[0]

	for _, version := range storage.Versions {
		if version.Atom == packageAtom {
			versions = append(versions, version)
		}
	}

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

// allRevisions returns all revisions of the given version
func allRevisions(versionSpecifier string, packageAtom string) []*models.Version {
	var versions []*models.Version
	revision := regexp.MustCompile(`-r[0-9]*$`)
	versionWithoutRevision := revision.Split(versionSpecifier, 1)[0]
	versionWithoutRevision = strings.ReplaceAll(versionWithoutRevision, "~", "")

	for _, version := range storage.Versions {
		if strings.HasPrefix(version.Id, versionWithoutRevision) {
			versions = append(versions, version)
		}
	}

	return versions
}

// exaktVersion returns the exact version specified in the versionSpecifier
func exaktVersion(versionSpecifier string, packageAtom string) []*models.Version {
	var versions []*models.Version

	for _, version := range storage.Versions {
		if version.Id == strings.Replace(versionSpecifier, "=", "", 1) {
			versions = append(versions, version)
		}
	}

	return versions
}

// TODO include subslot
// versionsWithSlot returns all versions with the given slot
func versionsWithSlot(versionSpecifier string, packageAtom string) []*models.Version {
	var versions []*models.Version
	slot := strings.Split(versionSpecifier, ":")[1]

	for _, version := range storage.Versions {
		if version.Atom == packageAtom && version.Slot == slot {
			versions = append(versions, version)
		}
	}

	return versions
}

// allVersions returns all versions of the given package
func allVersions(versionSpecifier string, packageAtom string) []*models.Version {
	var versions []*models.Version

	for _, version := range storage.Versions {
		if version.Atom == packageAtom {
			versions = append(versions, version)
		}
	}

	return versions
}

// maskVersions updates the MaskToVersion table using the given versions
func maskVersions(mask *models.Mask, versions []*models.Version) {

	for _, version := range versions {

		for _, storedVersion := range storage.Versions {
			if storedVersion.Id == version.Id {
				storedVersion.Masks = append(storedVersion.Masks, mask)
			}
		}

	}
}
