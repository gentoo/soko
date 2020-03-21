// Contains functions to import packages into the database

package repository

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"regexp"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"strings"
)

// isPackage checks whether the path points to a package
// descriptions that is an metadata.xml file
func isPackage(path string) bool {
	isPackage, _ := regexp.MatchString(`[^/]*\/[^/]*\/metadata\.xml`, path)
	return isPackage
}

// UpdatePackage updates the package in the database in case
// the given path points to a package description
func UpdatePackage(path string) {

	splittedLine := strings.Split(path, "\t")

	if len(splittedLine) != 2 {
		return
	}

	status := splittedLine[0]
	changedFile := splittedLine[1]

	if isPackage(changedFile) && status == "D" {
		updateDeletedPackage(changedFile)
	} else if isPackage(changedFile) && (status == "A" || status == "M") {
		updateModifiedPackage(changedFile)
	}
}

// updateDeletedPackage deletes a package from the database
func updateDeletedPackage(changedFile string) {
	splitted := strings.Split(changedFile, "/")
	category := splitted[0]
	packagename := splitted[1]
	atom := category + "/" + packagename

	gpackage := &models.Package{Atom: atom}
	_, err := database.DBCon.Model(gpackage).WherePK().Delete()

	if err != nil {
		logger.Error.Println("Error during deleting package " + atom)
		logger.Error.Println(err)
	}
}

// updateModifiedPackage adds a package to the database or
// updates it. To do so, it parses the metadata from metadata.xml
func updateModifiedPackage(changedFile string) {
	splitted := strings.Split(changedFile, "/")
	category := splitted[0]
	packagename := splitted[1]
	atom := category + "/" + packagename

	pkgmetadata := GetPkgMetadata(config.PortDir() + "/" + atom + "/metadata.xml")
	var maintainers []*models.Maintainer

	for _, maintainer := range pkgmetadata.MaintainerList {
		maintainer := &models.Maintainer{
			Name:     maintainer.Name,
			Type:     maintainer.Type,
			Email:    maintainer.Email,
			Restrict: maintainer.Restrict,
		}
		maintainers = append(maintainers, maintainer)
	}

	gpackage := &models.Package{
		Atom:            atom,
		Category:        category,
		Name:            packagename,
		Longdescription: pkgmetadata.Longdescription,
		Maintainers:     maintainers,
	}

	_, err := database.DBCon.Model(gpackage).OnConflict("(atom) DO UPDATE").
		Set("atom = EXCLUDED.atom").
		Set("category = EXCLUDED.category").
		Set("name = EXCLUDED.name").
		Set("longdescription = EXCLUDED.longdescription").
		Set("maintainers = EXCLUDED.maintainers").
		Insert()

	if err != nil {
		logger.Error.Println("Error during updating package " + atom)
		logger.Error.Println(err)
	}
}

// GetPkgMetadata reads and parses the package
// metadata from the metadata.xml file
func GetPkgMetadata(path string) Pkgmetadata {
	xmlFile, err := os.Open(path)
	if err != nil {
		logger.Error.Println("Error during reading package metadata")
		logger.Error.Println(err)
	}
	defer xmlFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlFile)
	var pkgmetadata Pkgmetadata
	xml.Unmarshal(byteValue, &pkgmetadata)
	return pkgmetadata
}

// Descriptions of the package metadata.xml format

type Pkgmetadata struct {
	XMLName         xml.Name     `xml:"pkgmetadata"`
	MaintainerList  []Maintainer `xml:"maintainer"`
	Longdescription string       `xml:"longdescription"`
}

type Maintainer struct {
	XMLName  xml.Name `xml:"maintainer"`
	Type     string   `xml:"type,attr"`
	Restrict string   `xml:"restrict,attr"`
	Email    string   `xml:"email"`
	Name     string   `xml:"name"`
}
