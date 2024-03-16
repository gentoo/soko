// Contains functions to import package versions into the database

package repository

import (
	"log/slog"
	"regexp"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"strings"
)

// isVersion checks whether the path points to a package version
// that is an .ebuild file
func isVersion(path string) bool {
	isVersion, _ := regexp.MatchString(`^[^/]*\/[^/]*\/.*\.ebuild$`, path)
	return isVersion
}

// UpdateVersions updates the versions in the database for each
// given path that points to a package version
func UpdateVersions(paths []string) {
	deleted := map[string]*models.Version{}
	modified := map[string]*models.Version{}

	for _, path := range paths {
		line := strings.Split(path, "\t")

		if len(line) != 2 {
			if len(line) == 1 && isVersion(path) {
				ver := updateModifiedVersion(path)
				modified[ver.Id] = ver
			}
			continue
		}

		status := line[0]
		changedFile := line[1]

		if !isVersion(changedFile) {
			continue
		} else if status == "D" {
			ver := updateDeletedVersion(changedFile)
			deleted[ver.Id] = ver
		} else if status == "A" || status == "M" {
			ver := updateModifiedVersion(changedFile)
			modified[ver.Id] = ver
		}
	}

	if len(deleted) > 0 {
		rows := make([]*models.Version, 0, len(deleted))
		for _, row := range deleted {
			rows = append(rows, row)
		}
		res, err := database.DBCon.Model(&rows).Delete()
		if err != nil {
			slog.Error("Failed deleting versions", slog.Any("err", err))
		} else {
			slog.Info("Deleted versions", slog.Int("rows", res.RowsAffected()))
		}
	}

	if len(modified) > 0 {
		rows := make([]*models.Version, 0, len(modified))
		for _, row := range modified {
			rows = append(rows, row)
		}
		res, err := database.DBCon.Model(&rows).OnConflict("(id) DO UPDATE").Insert()
		if err != nil {
			slog.Error("Failed updating versions", slog.Any("err", err))
		} else {
			slog.Info("Updated versions", slog.Int("rows", res.RowsAffected()))
		}
	}
}

// updateDeletedVersion deletes a package version from the database
func updateDeletedVersion(changedFile string) *models.Version {
	splitted := strings.Split(strings.TrimSuffix(changedFile, ".ebuild"), "/")
	category := splitted[0]
	packagename := splitted[1]
	version := strings.ReplaceAll(splitted[2], packagename+"-", "")

	atom := category + "/" + packagename
	id := atom + "-" + version

	return &models.Version{Id: id}
}

// updateModifiedVersion adds a package version to the database or
// updates it. To do so, it parses the metadata from the md5-cache
func updateModifiedVersion(changedFile string) *models.Version {
	splitted := strings.Split(strings.TrimSuffix(changedFile, ".ebuild"), "/")
	category := splitted[0]
	packagename := splitted[1]
	version := strings.ReplaceAll(splitted[2], packagename+"-", "")

	atom := category + "/" + packagename
	id := atom + "-" + version

	version_metadata, _ := utils.ReadLines(config.PortDir() + "/metadata/md5-cache/" + id)

	slot := "0"
	subslot := "0"
	var eapi, keywords, license, description string
	var useflags, restricts, properties, homepages []string

	for _, metadata := range version_metadata {
		switch {
		case strings.HasPrefix(metadata, "EAPI="):
			eapi = strings.TrimPrefix(metadata, "EAPI=")

		case strings.HasPrefix(metadata, "KEYWORDS="):
			keywords = strings.TrimPrefix(metadata, "KEYWORDS=")

		case strings.HasPrefix(metadata, "IUSE="):
			useflags = strings.Split(strings.TrimPrefix(metadata, "IUSE="), " ")

		case strings.HasPrefix(metadata, "RESTRICT="):
			restricts = strings.Split(strings.ReplaceAll(strings.TrimPrefix(metadata, "RESTRICT="), "!test? ( test )", ""), " ")
			if len(restricts) == 1 && restricts[0] == "" {
				restricts = []string{}
			}

		case strings.HasPrefix(metadata, "PROPERTIES="):
			properties = strings.Split(strings.TrimPrefix(metadata, "PROPERTIES="), " ")

		case strings.HasPrefix(metadata, "HOMEPAGE="):
			homepages = strings.Split(strings.TrimPrefix(metadata, "HOMEPAGE="), " ")

		case strings.HasPrefix(metadata, "LICENSE="):
			license = strings.TrimPrefix(metadata, "LICENSE=")

		case strings.HasPrefix(metadata, "DESCRIPTION="):
			description = strings.TrimPrefix(metadata, "DESCRIPTION=")

		case strings.HasPrefix(metadata, "SLOT="):
			rawSlot := strings.TrimPrefix(metadata, "SLOT=")
			slot, subslot, _ = strings.Cut(rawSlot, "/")
		}
	}

	return &models.Version{
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
}
