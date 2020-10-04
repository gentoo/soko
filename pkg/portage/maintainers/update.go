package maintainers

import (
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"soko/pkg/utils"
	"strings"
	"time"
)

func FullImport() {

	database.Connect()
	defer database.DBCon.Close()

	var allMaintainerInformation []*models.Maintainer
	database.DBCon.Model((*models.Package)(nil)).ColumnExpr("jsonb_array_elements(maintainers)->>'Name' as name, jsonb_array_elements(maintainers) ->> 'Email' as email, jsonb_array_elements(maintainers) ->> 'Type' as type").Select(&allMaintainerInformation)

	maintainers := map[string]*models.Maintainer{}

	for _, rawMaintainer := range allMaintainerInformation {
		_, ok := maintainers[rawMaintainer.Email]
		if !ok {
			maintainers[rawMaintainer.Email] = rawMaintainer
		} else {
			if maintainers[rawMaintainer.Email].Name == "" {
				maintainers[rawMaintainer.Email].Name = rawMaintainer.Name
			}
		}
	}

	var gpackages []*models.Package
	database.DBCon.Model(&gpackages).
		Relation("Outdated").
		Relation("PullRequests").
		Relation("Bugs").
		Relation("Versions").
		Relation("Versions.Bugs").
		Relation("Versions.PkgCheckResults").
		Select()

	// TODO in future we want an incremental update here
	// but for now we delete everything and insert it again
	// this is currently acceptable as it takes less than 2 seconds
	deleteAllMaintainers()

	for _, maintainer := range maintainers {
		outdated := 0
		securityBugs := 0
		pullrequestIds := []string{}
		nonSecurityBugs := 0
		stableRequests := 0

		for _, gpackage := range gpackages {
			found := false
			for _, packageMaintainer := range gpackage.Maintainers {
				if packageMaintainer.Email == maintainer.Email {
					found = true
				}
			}

			if found {
				outdated = outdated + len(gpackage.Outdated)

				for _, pullRequest := range gpackage.PullRequests {
					pullrequestIds = append(pullrequestIds, string(pullRequest.Id))
				}

				for _, bug := range gpackage.AllBugs() {
					if bug.Component == "Vulnerabilities" {
						securityBugs++
					} else {
						nonSecurityBugs++
					}
				}

				// Find Stable Requests
				for _, version := range gpackage.Versions {
					for _, pkgcheckWarning := range version.PkgCheckResults {
						if pkgcheckWarning.Class == "StableRequest" {
							stableRequests++
						}
					}
				}

			}
		}

		maintainer.PackagesInformation = models.MaintainerPackagesInformation{
			Outdated:       outdated,
			PullRequests:   len(utils.Deduplicate(pullrequestIds)),
			Bugs:           nonSecurityBugs,
			SecurityBugs:   securityBugs,
			StableRequests: stableRequests,
		}

		if maintainer.Name == "" {
			maintainer.Name = strings.Title(strings.Split(maintainer.Email, "@")[0])
		}

		if maintainer.Type == "project" && strings.HasPrefix(maintainer.Name, "Gentoo ") {
			maintainer.Name = strings.Replace(maintainer.Name, "Gentoo ", "", 1)
		}

		if maintainer.Type == "person" {
			if strings.HasSuffix(maintainer.Email, "@gentoo.org") {
				maintainer.Type = "gentoo-developer"
			} else {
				maintainer.Type = "proxied-maintainer"
			}
		}

		database.DBCon.Model(maintainer).WherePK().OnConflict("(email) DO UPDATE").Insert()
	}

	updateStatus()
}

// deleteAllMaintainers deletes all entries in the maintainers table
func deleteAllMaintainers() {
	var maintainers []*models.Maintainer
	database.DBCon.Model(&maintainers).Select()
	for _, maintainer := range maintainers {
		database.DBCon.Model(maintainer).WherePK().Delete()
	}
}

func contains(element string, elements []string) bool {
	for _, el := range elements {
		if element == el {
			return true
		}
	}
	return false
}

func updateStatus(){
	database.DBCon.Model(&models.Application{
		Id:         "maintainers",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}