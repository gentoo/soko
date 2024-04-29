package anitya

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
)

// API documentation: https://release-monitoring.org/static/docs/api.html#get--api-v2-packages-

const itemsPerPage = 250

type ApiPackage struct {
	Name          string `json:"name"`
	Project       string `json:"project"`
	StableVersion string `json:"stable_version"`
	Version       string `json:"version"`
}

type ApiResponse struct {
	Items      []ApiPackage `json:"items"`
	TotalItems int          `json:"total_items"`
}

func UpdateAnitya() {
	anityaPackages, err := readAllResults()
	if err != nil {
		slog.Error("Failed fetching anitya data", slog.Any("err", err))
		return
	} else if len(anityaPackages) == 0 {
		slog.Error("No anitya packages found")
	}

	packagesMap := make(map[string]int, len(anityaPackages))
	packages := make([]*models.Package, len(anityaPackages))
	for i, p := range anityaPackages {
		packages[i] = &models.Package{Atom: p.Name}
		packagesMap[p.Name] = i
	}

	err = database.DBCon.Model(&packages).WherePK().Relation("Versions").Select()
	if err != nil {
		slog.Error("Failed fetching packages", slog.Any("err", err))
		return
	}

	outdatedEntries := make([]*models.OutdatedPackages, 0, len(packages))

nextPackage:
	for _, p := range packages {
		anitya := anityaPackages[packagesMap[p.Atom]]
		p.AnityaInfo = &models.AnityaInfo{
			Project: anitya.Project,
		}
		if len(p.Versions) == 0 {
			continue
		}

		latest := models.Version{Version: anitya.LatestVersion()}
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
		outdatedEntries = append(outdatedEntries, &models.OutdatedPackages{
			Atom:          p.Atom,
			GentooVersion: currentLatest.Version,
			NewestVersion: anitya.StableVersion,
			Source:        models.OutdatedSourceAnitya,
		})
	}
	_, err = database.DBCon.Model(&packages).Set("anitya_info = ?anitya_info").Update()
	if err != nil {
		slog.Error("Failed updating packages", slog.Any("err", err))
		return
	}
	slog.Info("Updated anitya information", slog.Int("count", len(packages)))

	_, _ = database.DBCon.Model((*models.OutdatedPackages)(nil)).Where("source = ?", models.OutdatedSourceAnitya).Delete()
	res, err := database.DBCon.Model(&outdatedEntries).OnConflict("(atom) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Error while inserting outdated packages", slog.Any("err", err))
	} else {
		slog.Info("Inserted outdated packages", slog.Int("res", res.RowsAffected()))
	}

	updateStatus()
}

var client = http.Client{Timeout: 1 * time.Minute}

func fetchResults(page int, params url.Values) (int, []ApiPackage, error) {
	req, err := http.NewRequest("GET", "https://release-monitoring.org/api/v2/packages/?"+params.Encode(), nil)
	if err != nil {
		slog.Error("Failed creating request", slog.Int("page", page), slog.Any("err", err))
		return 0, nil, err
	}
	req.Header.Set("User-Agent", config.UserAgent())

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed fetching anitya data", slog.Int("page", page), slog.Any("err", err))
		return 0, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed fetching anitya data", slog.Int("page", page), slog.Int("status", resp.StatusCode))
		return 0, nil, nil
	}

	var data ApiResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	return data.TotalItems, data.Items, err
}

func readAllResults() (result []ApiPackage, err error) {
	params := url.Values{
		"distribution":   {"Gentoo"},
		"items_per_page": {strconv.Itoa(itemsPerPage)},
	}
	totalPages := 1
	for page := 1; page <= totalPages; page++ {
		slog.Info("Fetching anitya data", slog.Int("page", page))
		params.Set("page", strconv.Itoa(page))
		total, items, err := fetchResults(page, params)
		if err != nil {
			return nil, err
		}

		if page == 1 {
			totalPages = (total + itemsPerPage - 1) / itemsPerPage
			result = make([]ApiPackage, 0, total)
		}
		result = append(result, items...)
	}
	return
}

func (p *ApiPackage) LatestVersion() (result string) {
	result = p.Version
	if p.StableVersion != "" {
		result = p.StableVersion
	}
	result, _, _ = strings.Cut(result, ".post")
	return
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "anitya",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
