// Update the portage data in the database

package portage

import (
	"io/ioutil"
	"log"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
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

	if config.Quiet() == "true" {
		log.SetOutput(ioutil.Discard)
	}

	logger.Info.Println("Start update...")

	// update the local useflags
	repository.UpdateUse("profiles/use.local.desc")

	updateMetadata()
	updatePackageData()
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
//
// It works incrementally so that files are only parsed and updated whenever the
// file has been modified within the new commits. New commits are determined by
// retrieving the last commit in the database (if present) and parsing all
// following commits. In case no last commit is present a full import
// starting with the first commit in the tree is done.
func updateMetadata() {

	logger.Info.Println("Start updating changed metadata")

	latestCommit := utils.GetLatestCommit()

	changed := utils.ChangedFiles(latestCommit, "HEAD")
	logger.Info.Println("Iterating", len(changed), "changed files")
	for _, path := range changed {
		repository.UpdateUse(path)
		repository.UpdateMask(path)
		repository.UpdatePackagesDeprecated(path)
		repository.UpdateArch(path)
	}

}

// updatePackageData incrementally updates all package data in the database, that has
// been changed since the last update. That is:
//   - categories
//   - packages
//   - versions
//
// changed data is determined by parsing all commits since the last update.
func updatePackageData() {

	logger.Info.Println("Start updating changed package data")

	latestCommit := utils.GetLatestCommit()

	changed := utils.ChangedFiles(latestCommit, "HEAD")
	logger.Info.Println("Iterating", len(changed), "changed files")
	repository.UpdateVersions(changed)
	repository.UpdatePackages(changed)
	repository.UpdateCategories(changed)

}

// updateHistory incrementally imports all new commits. New commits are
// determined by retrieving the last commit in the database (if present)
// and parsing all following commits. In case no last commit is present
// a full import starting with the first commit in the tree is done.
func updateHistory() {

	logger.Info.Println("Start updating the history")

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

	_, err := database.DBCon.Model(application).
		OnConflict("(id) DO UPDATE").
		Insert()

	if err != nil {
		logger.Error.Println("Updating application data failed")
		logger.Error.Println(err)
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

	if config.Quiet() == "true" {
		log.SetOutput(ioutil.Discard)
	}

	logger.Info.Println("Full update up...")

	// Add new entries & update existing
	logger.Info.Println("Update all present files")

	// update the local useflags
	repository.UpdateUse("profiles/use.local.desc")

	allFiles := utils.AllFiles()
	repository.UpdateVersions(allFiles)
	repository.UpdatePackages(allFiles)
	repository.UpdateCategories(allFiles)

	// Delete removed entries
	logger.Info.Println("Delete removed files from the database")
	deleteRemovedVersions()
	deleteRemovedPackages()
	deleteRemovedCategories()

	fixPrecedingCommitsOfPackages()

	logger.Info.Println("Finished update up...")
}

// deleteRemovedVersions removes all versions from the database
// that are present in the database but not in the main tree.
func deleteRemovedVersions() {
	var versions, toDelete []*models.Version
	database.DBCon.Model(&versions).Select()

	for _, version := range versions {
		path := config.PortDir() + "/" + version.Atom + "/" + version.Package + "-" + version.Version + ".ebuild"
		if !utils.FileExists(path) {
			logger.Error.Println("Found ebuild version in the database that does not exist at:", path)
			toDelete = append(toDelete, version)
		}
	}

	if len(toDelete) > 0 {
		res, err := database.DBCon.Model(&toDelete).Delete()
		if err != nil {
			logger.Error.Println("Error deleting versions", err)
		} else {
			logger.Info.Println("Deleted", res.RowsAffected(), "versions")
		}
	}
}

// deleteRemovedPackages removes all packages from the database
// that are present in the database but not in the main tree.
func deleteRemovedPackages() {
	var packages, toDelete []*models.Package
	database.DBCon.Model(&packages).Select()

	for _, pkg := range packages {
		path := config.PortDir() + "/" + pkg.Atom
		if !utils.FileExists(path) {
			logger.Error.Println("Found package in the database that does not exist at:", path)
			toDelete = append(toDelete, pkg)
		}
	}

	if len(toDelete) > 0 {
		res, err := database.DBCon.Model(&toDelete).Delete()
		if err != nil {
			logger.Error.Println("Error deleting packages", err)
		} else {
			logger.Info.Println("Deleted", res.RowsAffected(), "packages")
		}
	}
}

// deleteRemovedCategories removes all categories from the database
// that are present in the database but not in the main tree.
func deleteRemovedCategories() {
	var categories, toDelete []*models.Category
	database.DBCon.Model(&categories).Select()

	for _, category := range categories {
		path := config.PortDir() + "/" + category.Name
		if !utils.FileExists(path) {
			logger.Error.Println("Found category in the database that does not exist at:", path)
			toDelete = append(toDelete, category)
		}
	}

	if len(toDelete) > 0 {
		res, err := database.DBCon.Model(&toDelete).Delete()
		if err != nil {
			logger.Error.Println("Error deleting categories", err)
		} else {
			logger.Info.Println("Deleted", res.RowsAffected(), "categories")
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
	var packages []*models.Package
	database.DBCon.Model(&packages).Where("preceding_commits = 0").Select()
	if len(packages) == 0 {
		return
	}

	logger.Error.Println("Found", len(packages), "packages with preceding commits == 0. This should not happen. Fixing...")
	for _, pkg := range packages {
		pkg.PrecedingCommits = 1
	}
	database.DBCon.Model(&packages).Update()
}

// GetApplicationData is used to retrieve the
// application data from the database
func getApplicationData() models.Application {
	// Select user by primary key.
	applicationData := &models.Application{Id: "latest"}
	err := database.DBCon.Model(applicationData).WherePK().Select()
	if err != nil {
		logger.Error.Println("Error fetching application data")
		return models.Application{
			Id:         "latest",
			LastUpdate: time.Now(),
			LastCommit: "unknown",
			Version:    "unknown",
		}
	}
	return *applicationData
}
