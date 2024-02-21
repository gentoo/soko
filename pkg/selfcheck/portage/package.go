// Contains functions to import packages into the database

package repository

import (
	"encoding/xml"
	"io"
	"os"
	"regexp"
	"soko/pkg/config"
	"soko/pkg/logger"
	"soko/pkg/models"
	"soko/pkg/selfcheck/storage"
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
	if isPackage(path) {
		updateModifiedPackage(path)
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

	longDescription := ""
	for _, l := range pkgmetadata.LongdescriptionList {
		if l.Language == "" || l.Language == "en" {
			longDescription = l.Content
		}
	}

	remoteIds := []models.RemoteId{}
	for _, r := range pkgmetadata.Upstream.RemoteIds {
		remoteIds = append(remoteIds, models.RemoteId{
			Type: r.Type,
			Id:   r.Content,
		})
	}

	upstream := models.Upstream{
		RemoteIds: remoteIds,
		Doc:       pkgmetadata.Upstream.Doc,
		BugsTo:    pkgmetadata.Upstream.BugsTo,
		Changelog: pkgmetadata.Upstream.Changelog,
	}

	addPackage(&models.Package{
		Atom:            atom,
		Category:        category,
		Name:            packagename,
		Longdescription: longDescription,
		Maintainers:     maintainers,
		Upstream:        upstream,
	})

}

func addPackage(newPackage *models.Package) {
	found := false
	for _, p := range storage.Packages {
		if p.Atom == newPackage.Atom {
			found = true
			break
		}
	}
	if !found {
		storage.Packages = append(storage.Packages, newPackage)
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
	byteValue, _ := io.ReadAll(xmlFile)
	var pkgmetadata Pkgmetadata
	xml.Unmarshal(byteValue, &pkgmetadata)
	return pkgmetadata
}

// Descriptions of the package metadata.xml format

type Pkgmetadata struct {
	XMLName             xml.Name              `xml:"pkgmetadata"`
	MaintainerList      []Maintainer          `xml:"maintainer"`
	LongdescriptionList []LongdescriptionItem `xml:"longdescription"`
	Upstream            Upstream              `xml:"upstream"`
}

type Maintainer struct {
	XMLName  xml.Name `xml:"maintainer"`
	Type     string   `xml:"type,attr"`
	Restrict string   `xml:"restrict,attr"`
	Email    string   `xml:"email"`
	Name     string   `xml:"name"`
}

type LongdescriptionItem struct {
	XMLName  xml.Name `xml:"longdescription"`
	Content  string   `xml:",chardata"`
	Language string   `xml:"lang,attr"`
}

type Upstream struct {
	XMLName   xml.Name   `xml:"upstream"`
	RemoteIds []RemoteId `xml:"remote-id"`
	BugsTo    []string   `xml:"bugs-to"`
	Doc       []string   `xml:"doc"`
	Changelog []string   `xml:"changelog"`
}

type RemoteId struct {
	XMLName xml.Name `xml:"remote-id"`
	Type    string   `xml:"type,attr"`
	Content string   `xml:",chardata"`
}
