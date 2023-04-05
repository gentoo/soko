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

	blockedRepos := readBlockList("ignored-repositories")
	blockedCategories := readBlockList("ignored-categories")
	blockedPackages := readBlockList("ignored-packages")

	var outdatedVersions []*models.OutdatedPackages
	for packagename := range repoPackages {
		outdated := make(map[string]bool)
		currentVersion := make(map[string]string)
		var newestVersion string

		// get the gentoo atom name first
		gentooPackages := make(map[string]struct{})
		for _, v := range repoPackages[packagename] {
			if v.Repo == "gentoo" {
				gentooPackages[v.VisibleName] = struct{}{}
			}
		}
	mainLoop:
		for _, v := range repoPackages[packagename] {
			category, _, _ := strings.Cut(v.VisibleName, "/")
			if v.Repo == "gentoo" {
				if v.Status == "newest" {
					outdated[v.VisibleName] = false
				} else if v.Status == "outdated" &&
					!containsPrefix(blockedCategories, category) &&
					!containsPrefix(blockedPackages, v.VisibleName) {
					if _, found := outdated[v.VisibleName]; !found {
						outdated[v.VisibleName] = true
					}
					if latest, found := currentVersion[v.VisibleName]; found {
						current := models.Version{Version: v.Version}
						if current.GreaterThan(models.Version{Version: latest}) {
							currentVersion[v.VisibleName] = v.Version
						}
					} else {
						currentVersion[v.VisibleName] = v.Version
					}
				}
			} else if len(newestVersion) == 0 && v.Status == "newest" && !contains(blockedRepos, v.Repo) {
				for atom := range gentooPackages {
					if contains(blockedPackages, atom+"::"+v.Repo) {
						continue mainLoop
					}
				}
				newestVersion = v.Version
			}
		}

		if len(newestVersion) == 0 {
			continue
		}

		for atom, outdated := range outdated {
			if outdated && packagename[0] == byte(letter) {
				outdatedVersions = append(outdatedVersions, &models.OutdatedPackages{
					Atom:          atom,
					GentooVersion: currentVersion[atom],
					NewestVersion: newestVersion,
				})

				category, _, found := strings.Cut(atom, "/")
				if found {
					outdatedCategories[category]++
				}
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

// readBlockList parses a block list and returns a list of
// lines whereas comments as well as empty lines are ignored
func readBlockList(file string) map[string]struct{} {
	blocklist := make(map[string]struct{})
	resp, err := http.Get("https://gitweb.gentoo.org/sites/soko-metadata.git/plain/repology/" + file)
	if err != nil {
		return blocklist
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	rawBlocklist := buf.String()

	for _, line := range strings.Split(rawBlocklist, "\n") {
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			blocklist[line] = struct{}{}
		}
	}
	return blocklist
}

// contains returns true if the given list includes
// the given string. Otherwise false is returned.
func contains(list map[string]struct{}, item string) bool {
	_, found := list[item]
	return found
}

// contains returns true if the given string is a prefix
// of an item in the given list. Otherwise false is returned.
func containsPrefix(list map[string]struct{}, item string) bool {
	for i := range list {
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
