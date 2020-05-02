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

	updateMetadata()
	updatePackageData()
	updateHistory()

	repository.CalculateMaskedVersions()

}

// updateMetadata updates all USE flags, package masks and arches in the database
// by parsing:
//  - profiles/use.desc
//  - profiles/use.local.desc
//  - profiles/use.local.desc
//  - profiles/desc/*
//  - profiles/package.mask
//  - profiles/arch.list
//
// It works incrementally so that files are only parsed and updated whenever the
// file has been modified within the new commits. New commits are determined by
// retrieving the last commit in the database (if present) and parsing all
// following commits. In case no last commit is present a full import
// starting with the first commit in the tree is done.
func updateMetadata() {

	logger.Info.Println("Start updating changed metadata")

	latestCommit := utils.GetLatestCommit()

	for _, path := range utils.ChangedFiles(latestCommit, "HEAD") {
		repository.UpdateUse(path)
		repository.UpdateMask(path)
		repository.UpdateArch(path)
	}

}

// updatePackageData incrementally updates all package data in the database, that has
// been changed since the last update. That is:
//  - categories
//  - packages
//  - versions
// changed data is determined by parsing all commits since the last update.
func updatePackageData() {

	logger.Info.Println("Start updating changed package data")

	latestCommit := utils.GetLatestCommit()

	for _, path := range utils.ChangedFiles(latestCommit, "HEAD") {
		repository.UpdateVersion(path)
		repository.UpdatePackage(path)
		repository.UpdateCategory(path)
	}

}

// updateHistory incrementally imports all new commits. New commits are
// determined by retrieving the last commit in the database (if present)
// and parsing all following commits. In case no last commit is present
// a full import starting with the first commit in the tree is done.
func updateHistory() {

	logger.Info.Println("Start updating the history")

	latestCommit := repository.UpdateCommits()

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
	for _, path := range utils.AllFiles() {

		if strings.Contains(path, "net-misc/chrony/metadata.xml"){
			//repository.UpdateVersion(path)
			repository.UpdatePackage(path)
			//repository.UpdateCategory(path)
		}
	}

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
	var versions []*models.Version
	database.DBCon.Model(&versions).Select()

	for _, version := range versions {
		path := config.PortDir() + "/" + version.Atom + "/" + version.Package + "-" + version.Version + ".ebuild"
		if !utils.FileExists(path) {

			logger.Error.Println("Found ebuild version in the database that does not exist at:")
			logger.Error.Println(path)

			_, err := database.DBCon.Model(version).WherePK().Delete()

			if err != nil {
				logger.Error.Println("Error deleting version " + version.Atom + " - " + version.Version)
				logger.Error.Println(err)
			}
		}

	}
}

// deleteRemovedPackages removes all packages from the database
// that are present in the database but not in the main tree.
func deleteRemovedPackages() {
	var packages []*models.Package
	database.DBCon.Model(&packages).Select()

	for _, gpackage := range packages {
		path := config.PortDir() + "/" + gpackage.Atom
		if !utils.FileExists(path) {

			logger.Error.Println("Found package in the database that does not exist at:")
			logger.Error.Println(path)

			_, err := database.DBCon.Model(gpackage).WherePK().Delete()

			if err != nil {
				logger.Error.Println("Error deleting package " + gpackage.Atom)
				logger.Error.Println(err)
			}
		}

	}
}

// deleteRemovedCategories removes all categories from the database
// that are present in the database but not in the main tree.
func deleteRemovedCategories() {
	var categories []*models.Category
	database.DBCon.Model(&categories).Select()

	for _, category := range categories {
		path := config.PortDir() + "/" + category.Name
		if !utils.FileExists(path) {

			logger.Error.Println("Found category in the database that does not exist at:")
			logger.Error.Println(path)

			_, err := database.DBCon.Model(category).WherePK().Delete()

			if err != nil {
				logger.Error.Println("Error deleting category " + category.Name)
				logger.Error.Println(err)
			}
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
	database.DBCon.Model(&packages).Select()
	for _, gpackage := range packages {
		if gpackage.PrecedingCommits == 0 {
			logger.Error.Println("Preceding Commits of package " + gpackage.Atom + " is null.")
			logger.Error.Println("This should not happen. Preceding Commits will be set to 1")
			gpackage.PrecedingCommits = 1
			database.DBCon.Model(gpackage).WherePK().Update()
		}
	}
}
