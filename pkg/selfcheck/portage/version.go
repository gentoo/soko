// SPDX-License-Identifier: GPL-2.0-only
// Contains functions to import package versions into the database

package repository

import (
	"regexp"
	"soko/pkg/config"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"soko/pkg/selfcheck/storage"
	"strings"
)

// isVersion checks whether the path points to a package version
// that is an .ebuild file
func isVersion(path string) bool {
	isVersion, _ := regexp.MatchString(`[^/]*\/[^/]*\/.*\.ebuild`, path)
	// ensures that files like /category/package/files/file.ebuild are NOT
	// recognized as version
	isVersion = isVersion && !strings.Contains(path, "/files/")
	return isVersion
}

// UpdateVersion updates the version in the database in case
// the given path points to a package version
func UpdateVersion(path string) {
	if isVersion(path) {
		updateModifiedVersion(path)
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

	version_metadata, _ := utils.ReadLines(config.SelfCheckPortDir() + "/metadata/md5-cache/" + id)

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
			restricts = strings.Split(strings.ReplaceAll(strings.ReplaceAll(metadata, "RESTRICT=", ""), "!test? ( test )", ""), " ")
			if len(restricts) == 1 && restricts[0] == "" {
				restricts = []string{}
			}

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

	addVersion(&models.Version{
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
	})

}

func addVersion(newVersion *models.Version) {
	found := false
	for _, v := range storage.Versions {
		if v.Id == newVersion.Id {
			found = true
			break
		}
	}
	if !found {
		storage.Versions = append(storage.Versions, newVersion)
	}
}
