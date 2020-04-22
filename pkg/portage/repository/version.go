// Contains functions to import package versions into the database

package repository

import (
	"regexp"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"strings"
)

// isVersion checks whether the path points to a package version
// that is an .ebuild file
func isVersion(path string) bool {
	isVersion, _ := regexp.MatchString(`[^/]*\/[^/]*\/.*\.ebuild`, path)
	return isVersion
}

// UpdateVersion updates the version in the database in case
// the given path points to a package version
func UpdateVersion(path string) {

	line := strings.Split(path, "\t")

	if len(line) != 2 {
		if len(line) == 1 && isVersion(path) {
			updateModifiedVersion(path)
		}
		return
	}

	status := line[0]
	changedFile := line[1]

	if isVersion(changedFile) && status == "D" {
		updateDeletedVersion(changedFile)
	} else if isVersion(changedFile) && (status == "A" || status == "M") {
		updateModifiedVersion(changedFile)
	}
}

// updateDeletedVersion deletes a package version from the database
func updateDeletedVersion(changedFile string) {
	splitted := strings.Split(strings.ReplaceAll(changedFile, ".ebuild", ""), "/")
	category := splitted[0]
	packagename := splitted[1]
	version := strings.ReplaceAll(splitted[2], packagename+"-", "")

	atom := category + "/" + packagename
	id := atom + "-" + version

	versionObject := &models.Version{Id: id}
	_, err := database.DBCon.Model(versionObject).WherePK().Delete()

	if err != nil {
		logger.Error.Println("Error during deleting version " + id)
		logger.Error.Println(err)
	}
}

// updateModifiedVersion adds a package version to the database or
// updates it. To do so, it parses the metadata from the md5-cache
func updateModifiedVersion(changedFile string) {
	splitted := strings.Split(strings.ReplaceAll(changedFile, ".ebuild", ""), "/")
	category := splitted[0]
	packagename := splitted[1]
	version := strings.ReplaceAll(splitted[2], packagename+"-", "")

	atom := category + "/" + packagename
	id := atom + "-" + version

	version_metadata, _ := utils.ReadLines(config.PortDir() + "/metadata/md5-cache/" + id)

	slot := "0"
	subslot := "0"
	eapi := ""
	keywords := ""
	var useflags []string
	var restricts []string
	var properties []string
	var homepages []string
	license := ""
	description := ""

	for _, metadata := range version_metadata {

		switch {
		case strings.HasPrefix(metadata, "EAPI="):
			eapi = strings.ReplaceAll(metadata, "EAPI=", "")

		case strings.HasPrefix(metadata, "KEYWORDS="):
			keywords = strings.ReplaceAll(metadata, "KEYWORDS=", "")

		case strings.HasPrefix(metadata, "IUSE="):
			useflags = strings.Split(strings.ReplaceAll(metadata, "IUSE=", ""), " ")

		case strings.HasPrefix(metadata, "RESTRICT="):
			restricts = strings.Split(strings.ReplaceAll(metadata, "RESTRICT=", ""), " ")

		case strings.HasPrefix(metadata, "PROPERTIES="):
			properties = strings.Split(strings.ReplaceAll(metadata, "PROPERTIES=", ""), " ")

		case strings.HasPrefix(metadata, "HOMEPAGE="):
			homepages = strings.Split(strings.ReplaceAll(metadata, "HOMEPAGE=", ""), " ")

		case strings.HasPrefix(metadata, "LICENSE="):
			license = strings.ReplaceAll(metadata, "LICENSE=", "")

		case strings.HasPrefix(metadata, "DESCRIPTION="):
			description = strings.ReplaceAll(metadata, "DESCRIPTION=", "")

		case strings.HasPrefix(metadata, "SLOT="):
			rawslot := strings.ReplaceAll(metadata, "SLOT=", "")
			slot = strings.Split(rawslot, "/")[0]
			if len(strings.Split(rawslot, "/")) > 1 {
				subslot = strings.Split(rawslot, "/")[1]
			}
		}

	}

	ebuildVersion := &models.Version{
		Id:          id,
		Category:    category,
		Package:     packagename,
		Atom:        atom,
		Version:     version,
		Slot:        slot,
		Subslot:     subslot,
		EAPI:        eapi,
		Keywords:    keywords,
		Useflags:    useflags,
		Restricts:   restricts,
		Properties:  properties,
		Homepage:    homepages,
		License:     license,
		Description: description,
	}

	_, err := database.DBCon.Model(ebuildVersion).OnConflict("(id) DO UPDATE").Insert()

	if err != nil {
		logger.Error.Println("Error during updating version " + id)
		logger.Error.Println(err)
	}
}
