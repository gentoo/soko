package repology

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"strings"
)

type Package struct {
	Repo        string `json:"repo"`
	Name        string `json:"name"`
	VisibleName string `json:"visiblename"`
	Version     string `json:"version"`
	Status      string `json:"status"`
}

type Packages map[string][]Package

// UpdateOutdated will update the database table that contains all outdated gentoo versions
func UpdateOutdated() {

	database.Connect()
	defer database.DBCon.Close()

	if config.Quiet() == "true" {
		log.SetOutput(ioutil.Discard)
	}

	// Get all outdated Versions
	var outdatedVersions []*models.OutdatedPackages
	letters := "abcdefghijklmnopqrstuvwxyz"
	for _, letter := range letters {
		outdatedVersions = append(outdatedVersions, getOutdatedStartingWith(letter)...)
	}

	// Clean up the database
	deleteAllOutdated()

	// Update the database
	for _, outdated := range outdatedVersions {
		database.DBCon.Insert(outdated)
	}
}

// getOutdatedStartingWith gets all outdated packages starting with the given letter
func getOutdatedStartingWith(letter rune) []*models.OutdatedPackages {
	repoPackages, err := parseRepologyData("https://repology.org/api/v1/projects/" + string(letter) + "/?inrepo=gentoo&outdated=1")

	if err != nil {
		logger.Error.Println("Error while fetching repology data")
		return []*models.OutdatedPackages{}
	}

	var outdatedVersions []*models.OutdatedPackages
	for packagename, _ := range repoPackages {
		atom := ""
		newest := ""
		version := ""
		outdated := false
		for _, v := range repoPackages[packagename] {
			if v.Status == "newest" {
				newest = v.Version
			}
			if v.Repo == "gentoo" && v.Status == "outdated" {
				atom = v.VisibleName
				outdated = true
				version = v.Version
			}
		}

		if outdated && strings.HasPrefix(packagename, string(letter)) {
			outdatedVersions = append(outdatedVersions, &models.OutdatedPackages{
				Atom:          atom,
				GentooVersion: version,
				NewestVersion: newest,
			})
		}
	}

	return outdatedVersions
}

// parseRepologyData gets the json from given url and parses it
func parseRepologyData(url string) (Packages, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Packages{}, err
	}
	defer resp.Body.Close()
	var repoPackages Packages
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&repoPackages)
	return repoPackages, err
}

// deleteAllOutdated deletes all entries in the outdated table
func deleteAllOutdated() {
	var allOutdated []*models.OutdatedPackages
	database.DBCon.Model(&allOutdated).Select()
	for _, outdated := range allOutdated {
		database.DBCon.Model(outdated).WherePK().Delete()
	}
}
