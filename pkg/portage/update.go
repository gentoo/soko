// Update the portage data in the database

package portage

import (
	"io/ioutil"
	"log"
	"soko/pkg/config"
	"soko/pkg/database"
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

	log.Println("Start update...")

	updateMetadata()
	updatePackageData()
	updateHistory()

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
func updateMetadata(){

	log.Print("Start updating changed metadata")

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
func updatePackageData(){

	log.Print("Start updating changed package data")

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
func updateHistory(){

	log.Print("Start updating the history")

	latestCommit := repository.UpdateCommits()

	application := &models.Application{
		Id:                "latest",
		LastUpdate:        time.Now(),
		Version:           config.Version(),
		LastCommit:        latestCommit,
	}

	_, err := database.DBCon.Model(application).
		OnConflict("(id) DO UPDATE").
		Insert()

	if err != nil {
		panic(err)
	}
}
