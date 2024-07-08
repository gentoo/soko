package repology

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

type Package struct {
	Repo        string `json:"repo"`
	Name        string `json:"name"`
	VisibleName string `json:"visiblename"`
	Version     string `json:"version"`
	Status      string `json:"status"`
}

type Packages = map[string][]Package

var client = http.Client{Timeout: 1 * time.Minute}
var clientRateLimiter = rate.NewLimiter(rate.Every(2*time.Second), 1)

// UpdateOutdated will update the database table that contains all outdated gentoo versions
func UpdateOutdated() {
	// Get all outdated Versions
	outdated := newOutdatedCheck()
	for letter := 'a'; letter <= 'z'; letter++ {
		outdated.getOutdatedStartingWith(letter)
	}
	outdated.compareAgainstVersions()

	// Update the database
	if len(outdated.outdatedVersions) > 0 {
		_, _ = database.DBCon.Model((*models.OutdatedPackages)(nil)).Where("source = ?", models.OutdatedSourceRepology).Delete()

		res, err := database.DBCon.Model(&outdated.outdatedVersions).OnConflict("(atom) DO NOTHING").Insert()
		if err != nil {
			slog.Error("Error while inserting outdated packages", slog.Any("err", err))
		} else {
			slog.Info("Inserted outdated packages", slog.Int("res", res.RowsAffected()))
		}
	}

	updateStatus()
}

type atomOutdatedRules struct {
	ignore         bool
	ignoreVersions []string
	ignoreRepos    []string
	selectedRepos  []string
}

func (a *atomOutdatedRules) isIgnored(version string, repo string) bool {
	if a == nil {
		return false
	} else if a.ignore {
		return true
	}

	for _, v := range a.ignoreVersions {
		if strings.HasPrefix(version, v) {
			return true
		}
	}
	for _, r := range a.ignoreRepos {
		if strings.HasPrefix(repo, r) {
			return true
		}
	}
	if len(a.selectedRepos) > 0 {
		found := false
		for _, r := range a.selectedRepos {
			if strings.HasPrefix(repo, r) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	return false
}

type outdatedCheck struct {
	blockedRepos      map[string]struct{}
	blockedCategories map[string]struct{}
	atomRules         map[string]*atomOutdatedRules

	outdatedVersions []*models.OutdatedPackages
}

func newOutdatedCheck() outdatedCheck {
	return outdatedCheck{
		blockedRepos:      readBlockList("ignored-repositories"),
		blockedCategories: readBlockList("ignored-categories"),
		atomRules:         buildAtomRules(),
	}
}

// getOutdatedStartingWith gets all outdated packages starting with the given letter
func (o *outdatedCheck) getOutdatedStartingWith(letter rune) {
	slog.Debug("Fetching outdated packages", slog.String("letter", string(letter)))
	repoPackages, err := parseRepologyData(letter)
	if err != nil {
		slog.Error("Error while fetching repology data", slog.String("letter", string(letter)), slog.Any("err", err))
	}

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
				} else if v.Status == "outdated" {
					if contains(o.blockedCategories, category) || o.atomRules[v.VisibleName].isIgnored(v.Version, v.Repo) {
						continue
					}
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
			} else if len(newestVersion) == 0 && v.Status == "newest" && !contains(o.blockedRepos, v.Repo) {
				for atom := range gentooPackages {
					if o.atomRules[atom].isIgnored(v.Version, v.Repo) {
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
				o.outdatedVersions = append(o.outdatedVersions, &models.OutdatedPackages{
					Atom:          atom,
					GentooVersion: currentVersion[atom],
					NewestVersion: newestVersion,
					Source:        models.OutdatedSourceRepology,
				})
			}
		}
	}
}

func (o *outdatedCheck) compareAgainstVersions() {
	packagesMap := make(map[string]*models.OutdatedPackages, len(o.outdatedVersions))
	packages := make([]*models.Package, len(o.outdatedVersions))
	for i, p := range o.outdatedVersions {
		packages[i] = &models.Package{Atom: p.Atom}
		packagesMap[p.Atom] = p
	}

	err := database.DBCon.Model(&packages).WherePK().Relation("Versions").Select()
	if err != nil {
		slog.Error("Failed fetching packages", slog.Any("err", err))
		return
	}

	o.outdatedVersions = make([]*models.OutdatedPackages, 0, len(packages))
nextPackage:
	for _, p := range packages {
		pkg := packagesMap[p.Atom]
		if len(p.Versions) == 0 {
			continue
		}

		latest := models.Version{Version: pkg.NewestVersion}
		if latest.Version == "" {
			continue
		}
		currentLatest := p.Versions[0]
		for _, v := range p.Versions {
			if slices.Contains(v.Properties, "live") {
				continue
			}
			if strings.HasPrefix(v.Version, latest.Version) || !latest.GreaterThan(*v) {
				continue nextPackage
			}
			if v.GreaterThan(*currentLatest) {
				currentLatest = v
			}
		}
		pkg.GentooVersion = currentLatest.Version
		o.outdatedVersions = append(o.outdatedVersions, pkg)
	}
	slog.Debug("Filtered outdated", slog.Int("before", len(packagesMap)), slog.Int("after", len(o.outdatedVersions)))
}

// parseRepologyData gets the json from given url and parses it
func parseRepologyData(letter rune) (Packages, error) {
	err := clientRateLimiter.Wait(context.Background())
	if err != nil {
		return Packages{}, fmt.Errorf("rate limiter failed: %w", err)
	}

	url := "https://repology.org/api/v1/projects/" + string(letter) + "/?inrepo=gentoo&outdated=1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Packages{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", config.UserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return Packages{}, fmt.Errorf("do http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Packages{}, fmt.Errorf("error while fetching repology data: %s", resp.Status)
	}

	var repoPackages Packages
	err = json.NewDecoder(resp.Body).Decode(&repoPackages)
	return repoPackages, err
}

// readBlockList parses a block list and returns a list of
// lines whereas comments as well as empty lines are ignored
func readBlockList(file string) map[string]struct{} {
	blocklist := make(map[string]struct{})
	resp, err := client.Get("https://gitweb.gentoo.org/sites/soko-metadata.git/plain/repology/" + file)
	if err != nil {
		slog.Error("Failed to fetch blacklist", slog.String("file", file), slog.Any("err", err))
		return blocklist
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to fetch blacklist", slog.String("file", file), slog.String("status", resp.Status))
		return blocklist
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			blocklist[line] = struct{}{}
		}
	}
	return blocklist
}

func buildAtomRules() map[string]*atomOutdatedRules {
	var versionNumber = regexp.MustCompile(`-[0-9]`)

	blacklist := readBlockList("ignored-packages")
	whitelist := readBlockList("selected-packages")

	atomRules := make(map[string]*atomOutdatedRules, len(blacklist)+len(whitelist))
	for line := range blacklist {
		cpv, repo, hasRepo := strings.Cut(line, "::")
		atom := versionNumber.Split(cpv, 2)[0]
		rule, found := atomRules[atom]
		if !found {
			rule = &atomOutdatedRules{}
			atomRules[atom] = rule
		}
		if hasRepo && repo != "" {
			rule.ignoreRepos = append(rule.ignoreRepos, repo)
		} else if atom != cpv {
			rule.ignoreVersions = append(rule.ignoreVersions, strings.TrimPrefix(cpv, atom+"-"))
		} else {
			rule.ignore = true
		}
	}

	for line := range whitelist {
		cpv, repo, hasRepo := strings.Cut(line, "::")
		atom := versionNumber.Split(cpv, 2)[0]
		rule, found := atomRules[atom]
		if !found {
			rule = &atomOutdatedRules{}
			atomRules[atom] = rule
		}
		if hasRepo && repo != "" {
			rule.selectedRepos = append(rule.selectedRepos, repo)
		}
	}

	return atomRules
}

// contains returns true if the given list includes
// the given string. Otherwise false is returned.
func contains(list map[string]struct{}, item string) bool {
	_, found := list[item]
	return found
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "repology",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
