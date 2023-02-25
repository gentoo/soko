package repology

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"strings"
	"time"
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
	outdatedCategories := make(map[string]int)
	var outdatedVersions []*models.OutdatedPackages
	letters := "abcdefghijklmnopqrstuvwxyz"
	for _, letter := range letters {
		outdatedVersions = append(outdatedVersions, getOutdatedStartingWith(letter, outdatedCategories)...)
	}

	// Clean up the database
	database.TruncateTable[models.OutdatedPackages]("atom")

	// Update the database
	if len(outdatedVersions) > 0 {
		database.DBCon.Model(&outdatedVersions).Insert()
	}

	// Updated the outdated status of categories
	var categories []*models.CategoryPackagesInformation
	err := database.DBCon.Model(&categories).Column("name").Select()
	if err != nil {
		logger.Error.Println("Error while fetching categories packages information", err)
		return
	} else if len(categories) > 0 {
		for _, category := range categories {
			category.Outdated = outdatedCategories[category.Name]
			delete(outdatedCategories, category.Name)
		}
		_, err = database.DBCon.Model(&categories).Set("outdated = ?outdated").Update()
		if err != nil {
			logger.Error.Println("Error while fetching categories packages information", err)
		}
		categories = make([]*models.CategoryPackagesInformation, 0, len(outdatedCategories))
	}

	for category, count := range outdatedCategories {
		categories = append(categories, &models.CategoryPackagesInformation{
			Name:     category,
			Outdated: count,
		})
	}
	if len(categories) > 0 {
		_, err = database.DBCon.Model(&categories).Insert()
		if err != nil {
			logger.Error.Println("Error while inserting categories packages information", err)
		}
	}

	updateStatus()
}

// getOutdatedStartingWith gets all outdated packages starting with the given letter
func getOutdatedStartingWith(letter rune, outdatedCategories map[string]int) []*models.OutdatedPackages {
	repoPackages, err := parseRepologyData("https://repology.org/api/v1/projects/" + string(letter) + "/?inrepo=gentoo&outdated=1")

	if err != nil {
		logger.Error.Println("Error while fetching repology data")
		return nil
	}

	blockedRepos := readBlocklist("ignored-repositories")
	blockedCategories := readBlocklist("ignored-categories")
	blockedPackages := readBlocklist("ignored-packages")

	var outdatedVersions []*models.OutdatedPackages
	for packagename := range repoPackages {
		var atom, newest, version string
		outdated := false
		// get the gentoo atom name first
		for _, v := range repoPackages[packagename] {
			if v.Repo == "gentoo" {
				atom = v.VisibleName
			}
		}
		for _, v := range repoPackages[packagename] {
			if v.Status == "newest" &&
				!contains(blockedRepos, v.Repo) &&
				!contains(blockedPackages, atom+"::"+v.Repo) {
				newest = v.Version
			}
			if v.Repo == "gentoo" && v.Status == "newest" {
				outdated = false
				break
			}
			if v.Repo == "gentoo" &&
				v.Status == "outdated" &&
				!containsPrefix(blockedCategories, strings.Split(v.VisibleName, "/")[0]) &&
				!containsPrefix(blockedPackages, v.VisibleName) {

				atom = v.VisibleName
				outdated = true
				version = v.Version
			}
		}

		if outdated && newest != "" && strings.HasPrefix(packagename, string(letter)) {
			outdatedVersions = append(outdatedVersions, &models.OutdatedPackages{
				Atom:          atom,
				GentooVersion: version,
				NewestVersion: newest,
			})

			category, _, found := strings.Cut(atom, "/")
			if found {
				outdatedCategories[category]++
			}
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
	err = json.NewDecoder(resp.Body).Decode(&repoPackages)
	return repoPackages, err
}

// readBlocklist parses a block list and returns a list of
// lines whereas comments as well as empty lines are ignored
func readBlocklist(file string) []string {
	resp, err := http.Get("https://gitweb.gentoo.org/sites/soko-metadata.git/plain/repology/" + file)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	rawBlocklist := buf.String()

	var blocklist []string
	for _, line := range strings.Split(rawBlocklist, "\n") {
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			blocklist = append(blocklist, line)
		}
	}
	return blocklist
}

// contains returns true if the given list includes
// the given string. Otherwise false is returned.
func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

// contains returns true if the given string is a prefix
// of an item in the given list. Otherwise false is returned.
func containsPrefix(list []string, item string) bool {
	for _, i := range list {
		if strings.HasPrefix(i, item) {
			return true
		}
	}
	return false
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "repology",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
