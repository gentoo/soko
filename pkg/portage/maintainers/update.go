// SPDX-License-Identifier: GPL-2.0-only
package maintainers

import (
	"log/slog"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const maintainerNeededEmail = "maintainer-needed@gentoo.org"

var caser = cases.Title(language.English)

func FullImport() {
	database.Connect()
	defer database.DBCon.Close()

	slog.Info("Loading all raw maintainers from the database")
	var allMaintainerInformation []*models.Maintainer
	_ = database.DBCon.Model((*models.Package)(nil)).ColumnExpr("jsonb_array_elements(maintainers)->>'Name' as name, jsonb_array_elements(maintainers) ->> 'Email' as email, jsonb_array_elements(maintainers) ->> 'Type' as type").Select(&allMaintainerInformation)

	maintainers := map[string]*models.Maintainer{
		maintainerNeededEmail: {
			Email: maintainerNeededEmail,
		},
	}

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

	slog.Info("Loading all packages from the database")
	var gpackages []*models.Package
	err := database.DBCon.Model(&gpackages).
		Relation("Outdated").
		Relation("PullRequests").
		Relation("Bugs").
		Relation("Versions").
		Relation("Versions.Bugs").
		Relation("Versions.PkgCheckResults").
		Select()
	if err != nil {
		slog.Error("Failed fetching packages", slog.Any("err", err))
		return
	}

	for _, maintainer := range maintainers {
		var outdated, stableRequests int
		pullRequestIds := make(map[string]struct{})
		maintainerPackages := []*models.Package{}

		for _, gpackage := range gpackages {
			found := false
			if len(gpackage.Maintainers) == 0 && maintainer.Email == maintainerNeededEmail {
				found = true
			} else {
				for _, packageMaintainer := range gpackage.Maintainers {
					if packageMaintainer.Email == maintainer.Email {
						found = true
					}
				}
			}

			if found {
				maintainerPackages = append(maintainerPackages, gpackage)

				outdated = outdated + len(gpackage.Outdated)

				for _, pullRequest := range gpackage.PullRequests {
					pullRequestIds[pullRequest.Id] = struct{}{}
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

		securityBugs, nonSecurityBugs := countBugs(maintainerPackages)

		maintainer.PackagesInformation = models.MaintainerPackagesInformation{
			Outdated:       outdated,
			PullRequests:   len(pullRequestIds),
			Bugs:           nonSecurityBugs,
			SecurityBugs:   securityBugs,
			StableRequests: stableRequests,
		}

		maintainer.Name = strings.TrimSpace(maintainer.Name)

		if maintainer.Name == "" {
			name, _, _ := strings.Cut(maintainer.Email, "@")
			maintainer.Name = caser.String(name)
		}

		if maintainer.Type == "project" && strings.HasPrefix(maintainer.Name, "Gentoo ") {
			maintainer.Name = strings.TrimPrefix(maintainer.Name, "Gentoo ")
		} else if maintainer.Type == "person" {
			if strings.HasSuffix(maintainer.Email, "@gentoo.org") {
				maintainer.Type = "gentoo-developer"
			} else {
				maintainer.Type = "proxied-maintainer"
			}
		}

	}

	// TODO in future we want an incremental update here
	// but for now we delete everything and insert it again
	// this is currently acceptable as it takes less than 2 seconds
	database.TruncateTable((*models.Maintainer)(nil))

	rows := make([]*models.Maintainer, 0, len(maintainers))
	for _, row := range maintainers {
		rows = append(rows, row)
	}
	res, err := database.DBCon.Model(&rows).OnConflict("(email) DO NOTHING").Insert()
	if err != nil {
		slog.Error("Failed inserting maintainers", slog.Any("err", err))
		return
	}
	slog.Info("Inserted maintainers", slog.Int("rows", res.RowsAffected()))

	updateStatus()
}

func countBugs(packages []*models.Package) (securityBugs, nonSecurityBugs int) {
	allBugs := make(map[string]*models.Bug)
	for _, gpackage := range packages {
		for _, bug := range gpackage.AllBugs() {
			allBugs[bug.Id] = bug
		}
	}

	for _, bug := range allBugs {
		if bug.Component == string(models.BugComponentVulnerabilities) {
			securityBugs++
		} else {
			nonSecurityBugs++
		}
	}

	return
}

func updateStatus() {
	_, err := database.DBCon.Model(&models.Application{
		Id:         "maintainers",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed updating status", slog.Any("err", err))
	}
}
