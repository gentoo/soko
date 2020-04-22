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


// CleanUp iterates through all ebuild versions in the database and
// checks whether they are still present in main tree. Normally there
// should be no version in the database anymore that is not present in
// the main gentoo tree. However in case this does happen, it does indiciate
// an error during the update process. In this case the version will be
// logged and deleted from the database. That is, CleanUp is currently
// used to a) find errors / outdated data and b) update the outdated data
// This method will be removed as soon as it shows that there are no
// errors present.
func CleanUp() {

	database.Connect()
	defer database.DBCon.Close()

	if config.Quiet() == "true" {
		log.SetOutput(ioutil.Discard)
	}

	logger.Info.Println("Start clean up...")

	var versions []*models.Version
	database.DBCon.Model(&versions).Select()

	for _, version := range versions {
		path := config.PortDir() + "/" + version.Atom + "/" + version.Package + "-" + version.Version + ".ebuild"
		if !utils.FileExists(path) {

			logger.Error.Println("Found ebuild version in the database that does not exist at:")
			logger.Error.Println(path)
			logger.Error.Println("The ebuild version got already deleted from the tree and should thus not exist in the database anymore:")
			logger.Error.Println(version.Atom + "-" + version.Version)

			_, err := database.DBCon.Model(version).WherePK().Delete()

			if err != nil {
				logger.Error.Println("Error deleting version")
				logger.Error.Println(version.Atom + " - " + version.Version)
				logger.Error.Println(err)
			}
		}

	}
	logger.Info.Println("Finished clean up...")
}
