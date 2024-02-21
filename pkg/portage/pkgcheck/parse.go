package pkgcheck

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"time"
)

// Descriptions of the xml format of the pkgcheck reports

type PkgCheckResults struct {
	XMLName xml.Name         `xml:"checks"`
	Results []PkgCheckResult `xml:"result"`
}

type PkgCheckResult struct {
	XMLName  xml.Name `xml:"result"`
	Category string   `xml:"category"`
	Package  string   `xml:"package"`
	Version  string   `xml:"version"`
	Class    string   `xml:"class"`
	Message  string   `xml:"msg"`
}

// UpdatePkgCheckResults will update the database table that contains all pkgcheck results
func UpdatePkgCheckResults() {

	database.Connect()
	defer database.DBCon.Close()

	if config.Quiet() == "true" {
		log.SetOutput(io.Discard)
	}

	// get the pkg check results from qa-reports.gentoo.org
	pkgCheckResults, err := parseQAReport()
	if err != nil {
		logger.Error.Println("Error while parsing qa-reports data. Aborting...")
		return
	}

	collected := make(map[string]*models.PkgCheckResult, len(pkgCheckResults))
	for _, pkgCheckResult := range pkgCheckResults {
		catpkg := pkgCheckResult.Category + "/" + pkgCheckResult.Package
		catpkgver := catpkg + "-" + pkgCheckResult.Version
		id := catpkgver + "-" + pkgCheckResult.Class + "-" + pkgCheckResult.Message
		collected[id] = &models.PkgCheckResult{
			Id:       id,
			Atom:     catpkg,
			Category: pkgCheckResult.Category,
			Package:  pkgCheckResult.Package,
			Version:  pkgCheckResult.Version,
			CPV:      catpkgver,
			Class:    pkgCheckResult.Class,
			Message:  pkgCheckResult.Message,
		}
	}

	// clean up the database
	database.TruncateTable[models.PkgCheckResult]("id")

	// update the database with the new results
	rows := make([]*models.PkgCheckResult, 0, len(collected))
	for _, row := range collected {
		rows = append(rows, row)
	}
	res, err := database.DBCon.Model(&rows).OnConflict("(id) DO NOTHING").Insert()
	if err != nil {
		logger.Error.Println("Error during inserting pkgcheck results", err)
		return
	}
	logger.Info.Println("Inserted", res.RowsAffected(), "pkgcheck results")

	updateCategoriesInfo()

	updateStatus()
}

// parseQAReport gets the xml from qa-reports.gentoo.org and parses it
func parseQAReport() ([]PkgCheckResult, error) {
	resp, err := http.Get("https://qa-reports.gentoo.org/output/gentoo-ci/output.xml")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var pkgCheckResults PkgCheckResults
	err = xml.NewDecoder(resp.Body).Decode(&pkgCheckResults)
	return pkgCheckResults.Results, err
}

func updateCategoriesInfo() {
	var categoriesInfoArr []*models.CategoryPackagesInformation
	err := database.DBCon.Model((*models.PkgCheckResult)(nil)).
		ColumnExpr("SPLIT_PART(atom, '/', 1) as name").
		ColumnExpr("COUNT(id) as stable_requests").
		Where("NULLIF(atom, '') IS NOT NULL").
		Where("class = 'StableRequest'").
		GroupExpr("SPLIT_PART(atom, '/', 1)").
		Select(&categoriesInfoArr)
	if err != nil {
		logger.Error.Println("Error while parsing qa-reports data. Aborting...", err)
		return
	}
	categoriesInfo := make(map[string]int, len(categoriesInfoArr))
	for _, categoryInfo := range categoriesInfoArr {
		categoriesInfo[categoryInfo.Name] = categoryInfo.StableRequests
	}

	var categories []*models.CategoryPackagesInformation
	err = database.DBCon.Model(&categories).Column("name").Select()
	if err != nil {
		logger.Error.Println("Error while fetching categories packages information", err)
		return
	} else if len(categories) > 0 {
		for _, category := range categories {
			category.StableRequests = categoriesInfo[category.Name]
			delete(categoriesInfo, category.Name)
		}
		_, err = database.DBCon.Model(&categories).Set("stable_requests = ?stable_requests").Update()
		if err != nil {
			logger.Error.Println("Error while fetching categories packages information", err)
		}
		categories = make([]*models.CategoryPackagesInformation, 0, len(categoriesInfo))
	}

	for category, stableRequests := range categoriesInfo {
		categories = append(categories, &models.CategoryPackagesInformation{
			Name:           category,
			StableRequests: stableRequests,
		})
	}
	if len(categories) > 0 {
		_, err = database.DBCon.Model(&categories).Insert()
		if err != nil {
			logger.Error.Println("Error while inserting categories packages information", err)
		}
	}
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "pkgcheck",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
