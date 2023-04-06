package utils

import (
	"regexp"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
)

var (
	revision = regexp.MustCompile(`-r[0-9]*$`)
)

func CalculateAffectedVersions(versionSpecifier, packageAtom string) []*models.Version {
	if strings.HasPrefix(versionSpecifier, "=") {
		return exactVersion(versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, "<=") {
		return comparedVersions("<=", versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, "<") {
		return comparedVersions("<", versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, ">=") {
		return comparedVersions(">=", versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, ">") {
		return comparedVersions(">", versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, "~") {
		return allRevisions(versionSpecifier, packageAtom)
	} else if strings.Contains(versionSpecifier, ":") {
		return versionsWithSlot(versionSpecifier, packageAtom)
	} else {
		return allVersions(versionSpecifier, packageAtom)
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
