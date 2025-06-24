// SPDX-License-Identifier: GPL-2.0-only
// Update the portage data in the database

package portage

import (
	"log/slog"
	"os"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"soko/pkg/portage/repository"
	"soko/pkg/portage/utils"
	"strings"
	"time"
)

// Update incrementally updates the whole data in the database. All commits
// since the last update are parsed and the changed data is updated. In case
// this is the first update that is there is no last update a full import
// starting with the first commit in the tree is done.
func Update() {
	database.Connect()
	defer database.DBCon.Close()

	slog.Info("Start update...")

	// update the local useflags
	repository.UpdateUse("profiles/use.local.desc")

	latestCommit := utils.GetLatestCommit()
	changed := utils.ChangedFiles(latestCommit, "HEAD")

	updateMetadata(changed)
	updatePackageData(changed)
	updateHistory()

	repository.CalculateMaskedVersions()
	repository.CalculateDeprecatedToVersion()
}

// updateMetadata updates all USE flags, package masks and arches in the database
// by parsing:
//   - profiles/use.desc
//   - profiles/use.local.desc
//   - profiles/use.local.desc
//   - profiles/desc/*
//   - profiles/package.mask
//   - profiles/package.deprecated
//   - profiles/arch.list
//   - profiles/updates/*
//
// It works incrementally so that files are only parsed and updated whenever the
// file has been modified within the new commits. New commits are determined by
// retrieving the last commit in the database (if present) and parsing all
// following commits. In case no last commit is present a full import
// starting with the first commit in the tree is done.
func updateMetadata(changed []string) {
	slog.Info("Start updating changed metadata")
	slog.Info("Iterating changed files", slog.Int("count", len(changed)))
	repository.UpdatePkgMoves(changed)
	for _, path := range changed {
		repository.UpdateUse(path)
		repository.UpdateMask(path)
		repository.UpdatePackagesDeprecated(path)
	}
}

// updatePackageData incrementally updates all package data in the database, that has
// been changed since the last update. That is:
//   - categories
//   - packages
//   - versions
//
// changed data is determined by parsing all commits since the last update.
func updatePackageData(changed []string) {
	slog.Info("Start updating changed package data")
	slog.Info("Iterating changed files", slog.Int("count", len(changed)))

	repository.UpdateVersions(changed)
	repository.UpdatePackages(changed)
	repository.UpdateCategories(changed)
}

// updateHistory incrementally imports all new commits. New commits are
// determined by retrieving the last commit in the database (if present)
// and parsing all following commits. In case no last commit is present
// a full import starting with the first commit in the tree is done.
func updateHistory() {
	slog.Info("Start updating the history")

	latestCommit := repository.UpdateCommits()

	if strings.TrimSpace(latestCommit) == "" {
		currentApplicationData := getApplicationData()
		latestCommit = currentApplicationData.LastCommit
	}

	application := &models.Application{
		Id:         "latest",
		LastUpdate: time.Now(),
		Version:    config.Version(),
		LastCommit: latestCommit,
	}

	_, err := database.DBCon.Model(application).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed updating application data", slog.Any("err", err))
	}
}

// FullUpdate does - as the name applies - a full update. That is, it
// iterates through *all* files in the tree and updates their records
// in the database. Afterwards it is checked whether all records that
// that are in the database, are present in the main tree. This way
// files which already got deleted from the main tree, get deleted
// from the database. All deleted files will be logged.
// This method is mainly intended for cleaning up the data and finding
// outdated data, which indicates bugs in the incremental update.
// Once there is no outdated data found anymore this method may become
// obsolete.
func FullUpdate() {
	database.Connect()
	defer database.DBCon.Close()

	slog.Info("Full update up...")

	// Add new entries & update existing
	slog.Info("Update all present files")

	// update useflags
	database.TruncateTable((*models.Useflag)(nil))
	repository.UpdateUse("profiles/use.desc")
	repository.UpdateUse("profiles/use.local.desc")
	if entries, err := os.ReadDir(config.PortDir() + "/profiles/desc"); err != nil {
		slog.Error("Error reading profiles/desc", slog.Any("err", err))
	} else {
		for _, entry := range entries {
			repository.UpdateUse("profiles/desc/" + entry.Name())
		}
	}

	allFiles := utils.AllFiles()
	updateMetadata(allFiles)
	repository.UpdateVersions(allFiles)
	repository.UpdatePackages(allFiles)
	repository.UpdateCategories(allFiles)

	// Delete removed entries
	slog.Info("Delete removed files from the database")
	deleteRemovedVersions()
	deleteRemovedPackages()
	deleteRemovedCategories()

	fixPrecedingCommitsOfPackages()

	repository.CalculateMaskedVersions()
	repository.CalculateDeprecatedToVersion()

	slog.Info("Finished update up...")
}

// deleteRemovedVersions removes all versions from the database
// that are present in the database but not in the main tree.
func deleteRemovedVersions() {
	var versions, toDelete []*models.Version
	err := database.DBCon.Model(&versions).Column("id", "atom", "package", "version").Select()
	if err != nil {
		slog.Error("Failed fetching versions", slog.Any("err", err))
		return
	}

	for _, version := range versions {
		path := config.PortDir() + "/" + version.Atom + "/" + version.Package + "-" + version.Version + ".ebuild"
		if !utils.FileExists(path) {
			slog.Error("Found ebuild version in the database that does not exist", slog.String("version", version.Id))
			toDelete = append(toDelete, version)
		}
	}

	if len(toDelete) > 0 {
		res, err := database.DBCon.Model(&toDelete).Delete()
		if err != nil {
			slog.Error("Failed deleting versions", slog.Any("err", err))
		} else {
			slog.Info("Deleted versions", slog.Int("rows", res.RowsAffected()))
		}
	}
}

// deleteRemovedPackages removes all packages from the database
// that are present in the database but not in the main tree.
func deleteRemovedPackages() {
	var packages, toDelete []*models.Package
	err := database.DBCon.Model(&packages).Column("atom").Select()
	if err != nil {
		slog.Error("Failed fetching packages", slog.Any("err", err))
		return
	}

	for _, pkg := range packages {
		if !utils.FileExists(config.PortDir() + "/" + pkg.Atom) {
			slog.Error("Found package in the database that does not exist", slog.String("atom", pkg.Atom))
			toDelete = append(toDelete, pkg)
		}
	}

	if len(toDelete) > 0 {
		res, err := database.DBCon.Model(&toDelete).Delete()
		if err != nil {
			slog.Error("Failed deleting packages", slog.Any("err", err))
		} else {
			slog.Info("Deleted packages", slog.Int("rows", res.RowsAffected()))
		}
	}
}

// deleteRemovedCategories removes all categories from the database
// that are present in the database but not in the main tree.
func deleteRemovedCategories() {
	var categories, toDelete []*models.Category
	err := database.DBCon.Model(&categories).Column("name").Select()
	if err != nil {
		slog.Error("Failed fetching categories", slog.Any("err", err))
		return
	}

	for _, category := range categories {
		if !utils.FileExists(config.PortDir() + "/" + category.Name) {
			slog.Error("Found category in the database that does not exist", slog.String("name", category.Name))
			toDelete = append(toDelete, category)
		}
	}

	if len(toDelete) > 0 {
		res, err := database.DBCon.Model(&toDelete).Delete()
		if err != nil {
			slog.Error("Failed deleting categories", slog.Any("err", err))
		} else {
			slog.Info("Deleted categories", slog.Int("rows", res.RowsAffected()))
		}
	}
}

// fixPreviousCommitsOfPackages updates packages that have
// preceding commits == null, that is preceding commits == 0
// This should not happen and will thus be logged. Furthermore
// preceding commits will be set to 1 in this case so that
// package does not mistakenly appears in the 'lately added
// packages' section.
func fixPrecedingCommitsOfPackages() {
	_, err := database.DBCon.Model((*models.Package)(nil)).
		Set("preceding_commits = 1").
		Where("preceding_commits = 0").
		Update()
	if err != nil {
		slog.Error("Failed updating packages with preceding commits == 0", slog.Any("err", err))
	}
}

// GetApplicationData is used to retrieve the
// application data from the database
func getApplicationData() models.Application {
	// Select user by primary key.
	applicationData := &models.Application{Id: "latest"}
	err := database.DBCon.Model(applicationData).WherePK().Select()
	if err != nil {
		slog.Error("Failed fetching application data", slog.Any("err", err))
		return models.Application{
			Id:         "latest",
			LastUpdate: time.Now(),
			LastCommit: "unknown",
			Version:    "unknown",
		}
	}
	return *applicationData
}
