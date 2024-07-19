// SPDX-License-Identifier: GPL-2.0-only
package bugs

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

type restAPIBug struct {
	Id                 int    `json:"id"`
	Product            string `json:"product"`
	Status             string `json:"status"`
	Summary            string `json:"summary"`
	Component          string `json:"component"`
	StabilizationAtoms string `json:"cf_stabilisation_atoms"`
	AssigneeDetails    struct {
		RealName string `json:"real_name"`
	} `json:"assigned_to_detail"`
}

func (b *restAPIBug) BugId() string {
	return strconv.Itoa(b.Id)
}

func (b *restAPIBug) ToDBType() *models.Bug {
	return &models.Bug{
		Id:        b.BugId(),
		Product:   b.Product,
		Status:    b.Status,
		Summary:   b.Summary,
		Component: b.Component,
		Assignee:  b.AssigneeDetails.RealName,
	}
}

func UpdateBugs() {
	database.Connect()
	defer database.DBCon.Close()

	update := models.Application{
		Id: "bugs",
	}
	err := database.DBCon.Model(&update).WherePK().Select()
	if err != nil && err != pg.ErrNoRows {
		slog.Error("Failed to fetch last update time for bugs", slog.Any("err", err))
		return
	}
	if update.LastCommit != "" {
		importAllOpenBugs()
	} else {
		lastUpdate := update.LastUpdate
		if err != nil {
			importAllOpenBugs()
		} else {
			if time.Now().Before(lastUpdate) {
				lastUpdate = time.Now()
			}
			changedSince := lastUpdate.AddDate(0, 0, -2)
			updateChangedBugs(changedSince)
		}
	}

	updateCategoriesInfo()

	updateStatus()
}

func fetchBugs(changedSince *time.Time, bugStatus []string) (bugs []restAPIBug, err error) {
	const limit = 5000

	params := url.Values{
		"include_fields": []string{"id,product,status,summary,component,assigned_to,cf_stabilisation_atoms"},
		"bug_status":     bugStatus,
		"order":          []string{"changeddate DESC"},
		"product":        []string{"Gentoo Linux", "Gentoo Security"},
		"limit":          []string{strconv.Itoa(limit)},
	}

	if changedSince != nil {
		params.Set("chfieldfrom", changedSince.Format("2006-01-02"))
	}

	for offset := 0; ; offset += limit {
		slog.Info("Importing bugs from bugs.gentoo.org", slog.Int("start", offset), slog.Int("end", offset+limit))
		params.Set("offset", strconv.Itoa(offset))
		resp, err := http.Get("https://bugs.gentoo.org/rest/bug?" + params.Encode())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			slog.Error("Failed to fetch bugs", slog.Int("status", resp.StatusCode))
			return bugs, nil
		}

		var response struct {
			Bugs []restAPIBug `json:"bugs"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			slog.Error("Failed to decode bugs", slog.Any("err", err))
			return bugs, nil
		}

		bugs = append(bugs, response.Bugs...)

		if len(response.Bugs) < limit {
			break
		}
	}
	slog.Info("Collected bugs", slog.Int("count", len(bugs)))
	return
}

func importAllOpenBugs() {
	bugs, err := fetchBugs(nil, []string{"UNCONFIRMED", "CONFIRMED", "IN_PROGRESS"})
	if err != nil {
		slog.Error("Failed to fetch bugs",
			slog.String("status", "UNCONFIRMED, CONFIRMED, IN_PROGRESS"),
			slog.Any("err", err))
		return
	}

	database.TruncateTable((*models.Bug)(nil))
	database.TruncateTable((*models.PackageToBug)(nil))
	database.TruncateTable((*models.VersionToBug)(nil))

	processApiBugs(bugs)
}

func updateChangedBugs(changedSince time.Time) {
	bugs, err := fetchBugs(&changedSince, []string{"UNCONFIRMED", "CONFIRMED", "IN_PROGRESS", "RESOLVED"})
	if err != nil {
		slog.Error("Failed to fetch bugs",
			slog.String("status", "UNCONFIRMED, CONFIRMED, IN_PROGRESS, RESOLVED"),
			slog.Time("changed_since", changedSince),
			slog.Any("err", err))
		return
	}
	processApiBugs(bugs)
}

func processApiBugs(bugs []restAPIBug) {
	var resolvedBugs []string
	var dbBugs []*models.Bug
	var verBugs []*models.VersionToBug
	var pkgsBugs []*models.PackageToBug
	processedBugs := make(map[int]struct{}, len(bugs))

	for _, bug := range bugs {
		if bug.Status == "RESOLVED" {
			resolvedBugs = append(resolvedBugs, bug.BugId())
		} else if _, found := processedBugs[bug.Id]; !found {
			dbBugs = append(dbBugs, bug.ToDBType())
			processedBugs[bug.Id] = struct{}{}
			bugId := bug.BugId()
			if strings.TrimSpace(bug.StabilizationAtoms) != "" {
				versions := make(map[string]struct{})
				for _, gpackage := range strings.Split(bug.StabilizationAtoms, "\n") {
					affectedVersions, _, _ := strings.Cut(strings.TrimSpace(gpackage), " ")
					if strings.TrimSpace(affectedVersions) != "" {
						for _, version := range calculateAffectedVersions(affectedVersions) {
							versions[version.Id] = struct{}{}
						}
					}
				}
				for version := range versions {
					verBugs = append(verBugs, &models.VersionToBug{
						Id:        version + "-" + bugId,
						VersionId: version,
						BugId:     bugId,
					})
				}
			} else {
				summary, _, _ := strings.Cut(strings.TrimSpace(bug.Summary), " ")
				affectedPackage := versionSpecifierToPackageAtom(summary)

				pkgsBugs = append(pkgsBugs, &models.PackageToBug{
					Id:          affectedPackage + "-" + bugId,
					PackageAtom: affectedPackage,
					BugId:       bugId,
				})
			}
		}
	}

	res1, err := database.DBCon.Model(&dbBugs).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed to insert bugs", slog.Any("err", err))
		return
	}

	res2, err := database.DBCon.Model(&verBugs).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed to insert version bugs", slog.Any("err", err))
		return
	}

	res3, err := database.DBCon.Model(&pkgsBugs).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed to insert package bugs", slog.Any("err", err))
		return
	}

	slog.Info("Inserted",
		slog.Int("bugs", res1.RowsAffected()),
		slog.Int("version_bugs", res2.RowsAffected()),
		slog.Int("package_bugs", res3.RowsAffected()))

	if len(resolvedBugs) > 0 {
		res1, err := database.DBCon.Model((*models.Bug)(nil)).WhereIn("id IN (?)", resolvedBugs).Delete()
		if err != nil {
			slog.Error("Failed to delete bugs", slog.Any("err", err))
			return
		}

		res2, err := database.DBCon.Model((*models.PackageToBug)(nil)).WhereIn("bug_id IN (?)", resolvedBugs).Delete()
		if err != nil {
			slog.Error("Failed to delete package bugs", slog.Any("err", err))
			return
		}

		res3, err := database.DBCon.Model((*models.VersionToBug)(nil)).WhereIn("bug_id IN (?)", resolvedBugs).Delete()
		if err != nil {
			slog.Error("Failed to delete version bugs", slog.Any("err", err))
			return
		}

		slog.Info("Deleted",
			slog.Int("bugs", res1.RowsAffected()),
			slog.Int("package_bugs", res2.RowsAffected()),
			slog.Int("version_bugs", res3.RowsAffected()))
	}
}

func calculateAffectedVersions(versionSpecifier string) []*models.Version {
	packageAtom := versionSpecifierToPackageAtom(versionSpecifier)
	return utils.CalculateAffectedVersions(versionSpecifier, packageAtom)
}

var versionNumber = regexp.MustCompile(`-[0-9]`)

// versionSpecifierToPackageAtom returns the package atom from a given version specifier
func versionSpecifierToPackageAtom(versionSpecifier string) string {
	gpackage := strings.ReplaceAll(versionSpecifier, ">", "")
	gpackage = strings.ReplaceAll(gpackage, "<", "")
	gpackage = strings.ReplaceAll(gpackage, "=", "")
	gpackage = strings.ReplaceAll(gpackage, "~", "")

	gpackage, _, _ = strings.Cut(gpackage, ":")

	gpackage = versionNumber.Split(gpackage, 2)[0]

	return gpackage
}

func updateCategoriesInfo() {
	var categoriesInfoArr []*models.CategoryPackagesInformation
	err := database.DBCon.Model((*models.PackageToBug)(nil)).
		ColumnExpr("SPLIT_PART(package_atom, '/', 1) as name").
		ColumnExpr("COUNT(DISTINCT bug_id) as bugs").
		ColumnExpr("COUNT(DISTINCT bug_id) FILTER(WHERE component = ?) as security_bugs", models.BugComponentVulnerabilities).
		Join("JOIN bugs").JoinOn("package_to_bug.bug_id = bugs.id").
		Where("NULLIF(package_atom, '') IS NOT NULL").
		Where(`package_atom LIKE '%/%'`).
		GroupExpr("SPLIT_PART(package_atom, '/', 1)").
		Select(&categoriesInfoArr)
	if err != nil {
		slog.Error("Failed collecting bugs stats", slog.Any("err", err))
		return
	}
	categoriesInfo := make(map[string]*models.CategoryPackagesInformation, len(categoriesInfoArr))
	for _, categoryInfo := range categoriesInfoArr {
		if categoryInfo.Name != "" {
			categoriesInfo[categoryInfo.Name] = categoryInfo
		}
	}

	var categories []*models.CategoryPackagesInformation
	err = database.DBCon.Model(&categories).Column("name").Select()
	if err != nil {
		slog.Error("Failed fetching categories packages information", slog.Any("err", err))
		return
	} else if len(categories) > 0 {
		for _, category := range categories {
			if info, found := categoriesInfo[category.Name]; found {
				category.Bugs = info.Bugs
				category.SecurityBugs = info.SecurityBugs
			} else {
				category.Bugs = 0
				category.SecurityBugs = 0
			}
			delete(categoriesInfo, category.Name)
		}
		_, err = database.DBCon.Model(&categories).
			Set("bugs = ?bugs").
			Set("security_bugs = ?security_bugs").
			Update()
		if err != nil {
			slog.Error("Failed updating categories packages information", slog.Any("err", err))
		}
		categories = make([]*models.CategoryPackagesInformation, 0, len(categoriesInfo))
	}

	for _, catInfo := range categoriesInfo {
		categories = append(categories, catInfo)
	}
	if len(categories) > 0 {
		_, err = database.DBCon.Model(&categories).Insert()
		if err != nil {
			slog.Error("Failed inserting categories packages information", slog.Any("err", err))
		}
	}
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "bugs",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
