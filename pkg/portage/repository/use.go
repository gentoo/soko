// Contains functions to import USE flags into the database

package repository

import (
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"strings"
)

// UpdateUse reads all USE flags descriptions from the given file in
// case the given file contains USE flags descriptions and imports
// each USE flag into the database
func UpdateUse(path string) {
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

	if status != "D" && (isLocalUseflag(changedFile) || isGlobalUseflag(changedFile) || isUseExpand(changedFile)) {

		rawFlags, _ := utils.ReadLines(config.PortDir() + "/" + changedFile)

		useFlags := make(map[string]*models.Useflag, len(rawFlags))
		for _, rawFlag := range rawFlags {
			if strings.TrimSpace(rawFlag) == "" || rawFlag[0] == '#' {
				continue
			}
			scope := getScope(changedFile)
			switch scope {
			case "local", "global":
				if flag := createUseflag(rawFlag, scope); flag != nil {
					useFlags[flag.Id] = flag
				}
			case "use_expand":
				file := strings.Split(changedFile, "/")[2]
				flag := createUseExpand(rawFlag, file)
				useFlags[flag.Id] = flag
			}
		}

		if len(useFlags) > 0 {
			rows := make([]*models.Useflag, 0, len(useFlags))
			for _, row := range useFlags {
				rows = append(rows, row)
			}
			res, err := database.DBCon.Model(&rows).OnConflict("(id) DO UPDATE").Insert()
			if err != nil {
				logger.Error.Println("Error during updating use flags", err)
			} else {
				logger.Info.Println("Updated", res.RowsAffected(), "use flags")
			}
		}
	}

}

// createUseflag parses the description from the file,
// creates a USE flag and imports it into the database
func createUseflag(rawFlag string, scope string) *models.Useflag {
	pkguse, description, found := strings.Cut(rawFlag, " - ")
	if !found {
		return nil
	}

	pkg, use, found := strings.Cut(pkguse, ":")
	if found != (scope == "local") {
		return nil
	} else if !found {
		use = pkguse
	}

	return &models.Useflag{
		Id:          pkguse + "-" + scope,
		Package:     pkg,
		Name:        use,
		Scope:       scope,
		Description: description,
	}
}

// createUseExpand parses the description from the file,
// creates a USE expand flag and imports it into the database
func createUseExpand(rawFlag string, file string) *models.Useflag {
	group := strings.TrimSuffix(file, ".desc")
	unexpanded, description, _ := strings.Cut(rawFlag, " - ")
	id := group + "_" + unexpanded

	return &models.Useflag{
		Id:          id,
		Name:        id,
		Scope:       "use_expand",
		Description: description,
		UseExpand:   group,
	}
}

// getScope returns either "local", "global", "use_expand"
// or "" based on the file that the path points to
func getScope(path string) string {
	if isLocalUseflag(path) {
		return "local"
	} else if isGlobalUseflag(path) {
		return "global"
	} else if isUseExpand(path) {
		return "use_expand"
	}
	return ""
}

// isLocalUseflag checks whether the path points to
// the file that contains the local USE flags
func isLocalUseflag(path string) bool {
	return path == "profiles/use.local.desc"
}

// isGlobalUseflag checks whether the path points to
// the file that contains the global USE flags
func isGlobalUseflag(path string) bool {
	return path == "profiles/use.desc"
}

// isGlobalUseflag checks whether the path points to
// a file that contains use expand flags
func isUseExpand(path string) bool {
	return strings.HasPrefix(path, "profiles/desc/")
}
