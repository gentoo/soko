package pkgcheck

import (
	"encoding/xml"
	"io/ioutil"
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
		log.SetOutput(ioutil.Discard)
	}

	// get the pkg check results from qa-reports.gentoo.org
	pkgCheckResults, err := parseQAReport()
	if err != nil {
		logger.Error.Println("Error while parsing qa-reports data. Aborting...")
	}

	// clean up the database
	deleteAllPkgCheckResults()

	// update the database with the new results
	for _, pkgCheckResult := range pkgCheckResults.Results {
		database.DBCon.Insert(&models.PkgCheckResult{
			Id:       pkgCheckResult.Category + "/" + pkgCheckResult.Package + "-" + pkgCheckResult.Version + "-" + pkgCheckResult.Class + "-" + pkgCheckResult.Message,
			Atom:     pkgCheckResult.Category + "/" + pkgCheckResult.Package,
			Category: pkgCheckResult.Category,
			Package:  pkgCheckResult.Package,
			Version:  pkgCheckResult.Version,
			CPV:      pkgCheckResult.Category + "/" + pkgCheckResult.Package + "-" + pkgCheckResult.Version,
			Class:    pkgCheckResult.Class,
			Message:  pkgCheckResult.Message,
		})
	}

	updateStatus()
}

// parseQAReport gets the xml from qa-reports.gentoo.org and parses it
func parseQAReport() (PkgCheckResults, error) {
	resp, err := http.Get("https://qa-reports.gentoo.org/output/gentoo-ci/output.xml")
	if err != nil {
		return PkgCheckResults{}, err
	}
	defer resp.Body.Close()
	xmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PkgCheckResults{}, err
	}
	var pkgCheckResults PkgCheckResults
	xml.Unmarshal(xmlData, &pkgCheckResults)
	return pkgCheckResults, err
}

// deleteAllOutdated deletes all entries in the outdated table
func deleteAllPkgCheckResults() {
	var allPkgCheckResults []*models.PkgCheckResult
	database.DBCon.Model(&allPkgCheckResults).Select()
	for _, pkgCheckResult := range allPkgCheckResults {
		database.DBCon.Model(pkgCheckResult).WherePK().Delete()
	}
}

func updateStatus(){
	database.DBCon.Model(&models.Application{
		Id:         "pkgcheck",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}