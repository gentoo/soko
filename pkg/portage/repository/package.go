// Contains functions to import packages into the database

package repository

import (
	"encoding/xml"
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
	isPackage, _ := regexp.MatchString(`^[^/]*\/[^/]*\/metadata\.xml$`, path)
	return isPackage
}

// UpdatePackages updates the packages in the database for each
// given path that points to a package description
func UpdatePackages(paths []string) {
	deleted := map[string]*models.Package{}
	modified := map[string]*models.Package{}

	for _, path := range paths {
		splittedLine := strings.Split(path, "\t")

		if len(splittedLine) != 2 {
			if len(splittedLine) == 1 && isPackage(path) {
				if pkg := updateModifiedPackage(path); pkg != nil {
					modified[pkg.Atom] = pkg
				}
			}
			continue
		}

		status := splittedLine[0]
		changedFile := splittedLine[1]

		if !isPackage(changedFile) {
			continue
		} else if status == "D" {
			pkg := updateDeletedPackage(changedFile)
			deleted[pkg.Atom] = pkg
		} else if status == "A" || status == "M" {
			if pkg := updateModifiedPackage(changedFile); pkg != nil {
				modified[pkg.Atom] = pkg
			}
		}
	}

	if len(deleted) > 0 {
		rows := make([]*models.Package, 0, len(deleted))
		for _, row := range deleted {
			rows = append(rows, row)
		}
		res, err := database.DBCon.Model(&rows).Delete()
		if err != nil {
			logger.Error.Println("Error during deleting packages", err)
		} else {
			logger.Info.Println("Deleted", res.RowsAffected(), "packages")
		}
	}

	if len(modified) > 0 {
		rows := make([]*models.Package, 0, len(modified))
		for _, row := range modified {
			rows = append(rows, row)
		}
		res, err := database.DBCon.Model(&rows).OnConflict("(atom) DO UPDATE").
			Set("atom = EXCLUDED.atom").
			Set("category = EXCLUDED.category").
			Set("name = EXCLUDED.name").
			Set("longdescription = EXCLUDED.longdescription").
			Set("maintainers = EXCLUDED.maintainers").
			Set("upstream = EXCLUDED.upstream").
			Insert()
		if err != nil {
			logger.Error.Println("Error during updating packages", err)
		} else {
			logger.Info.Println("Updated", res.RowsAffected(), "packages")
		}
	}
}

// updateDeletedPackage deletes a package from the database
func updateDeletedPackage(changedFile string) *models.Package {
	splitted := strings.Split(changedFile, "/")
	category := splitted[0]
	packagename := splitted[1]
	atom := category + "/" + packagename

	return &models.Package{Atom: atom}
}

// updateModifiedPackage adds a package to the database or
// updates it. To do so, it parses the metadata from metadata.xml
func updateModifiedPackage(changedFile string) *models.Package {
	splitted := strings.Split(changedFile, "/")
	category := splitted[0]
	packagename := splitted[1]
	atom := category + "/" + packagename

	xmlFile, err := os.Open(config.PortDir() + "/" + atom + "/metadata.xml")
	if err != nil {
		logger.Error.Println("Error during reading package metadata", err)
		return nil
	}
	defer xmlFile.Close()
	var pkgMetadata PkgMetadata
	err = xml.NewDecoder(xmlFile).Decode(&pkgMetadata)
	if err != nil {
		logger.Error.Println("Error during package", changedFile, "decoding", err)
		return nil
	}

	maintainers := make([]*models.Maintainer, len(pkgMetadata.MaintainerList))
	for i, maintainer := range pkgMetadata.MaintainerList {
		maintainers[i] = &models.Maintainer{
			Name:     maintainer.Name,
			Type:     maintainer.Type,
			Email:    maintainer.Email,
			Restrict: maintainer.Restrict,
		}
	}

	var longDescription string
	for _, l := range pkgMetadata.LongDescriptionList {
		if l.Language == "" || l.Language == "en" {
			longDescription = l.Content
		}
	}

	remoteIds := make([]models.RemoteId, len(pkgMetadata.Upstream.RemoteIds))
	for i, r := range pkgMetadata.Upstream.RemoteIds {
		remoteIds[i] = models.RemoteId{
			Type: r.Type,
			Id:   r.Content,
		}
	}

	upstream := models.Upstream{
		RemoteIds: remoteIds,
		Doc:       pkgMetadata.Upstream.Doc,
		BugsTo:    pkgMetadata.Upstream.BugsTo,
		Changelog: pkgMetadata.Upstream.Changelog,
	}

	return &models.Package{
		Atom:            atom,
		Category:        category,
		Name:            packagename,
		Longdescription: longDescription,
		Maintainers:     maintainers,
		Upstream:        upstream,
	}
}

// Descriptions of the package metadata.xml format

type PkgMetadata struct {
	XMLName             xml.Name              `xml:"pkgmetadata"`
	MaintainerList      []Maintainer          `xml:"maintainer"`
	LongDescriptionList []LongDescriptionItem `xml:"longdescription"`
	Upstream            Upstream              `xml:"upstream"`
}

type Maintainer struct {
	XMLName  xml.Name `xml:"maintainer"`
	Type     string   `xml:"type,attr"`
	Restrict string   `xml:"restrict,attr"`
	Email    string   `xml:"email"`
	Name     string   `xml:"name"`
}

type LongDescriptionItem struct {
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
