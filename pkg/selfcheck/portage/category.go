// Contains functions to import categories into the database

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

// isCategory checks whether the path points to a category
// descriptions that is an metadata.xml file
func isCategory(path string) bool {
	isCategory, _ := regexp.MatchString(`[^/]*\/metadata\.xml`, path)
	return isCategory
}

// UpdateCategory updates the category in the database in case
// the given path points to a category description
func UpdateCategory(path string) {
	if isCategory(path) {
		updateModifiedCategory(path)
	}
}

// updateModifiedCategory adds a category to the database or
// updates it. To do so, it parses the metadata from metadata.xml
func updateModifiedCategory(changedFile string) {
	splitted := strings.Split(changedFile, "/")
	id := splitted[0]

	catmetadata := GetCatMetadata(config.PortDir() + "/" + changedFile)
	description := ""

	for _, longdescription := range catmetadata.Longdescriptions {
		if longdescription.Lang == "en" {
			description = strings.TrimSpace(longdescription.Content)
		}
	}

	addCategory(&models.Category{
		Name:        id,
		Description: description,
	})
}

func addCategory(category *models.Category) {
	found := false
	for _, cat := range storage.Categories {
		if cat.Name == category.Name {
			found = true
			break
		}
	}
	if !found {
		storage.Categories = append(storage.Categories, category)
	}
}

// GetCatMetadata reads and parses the category
// metadata from the metadata.xml file
func GetCatMetadata(path string) Catmetadata {
	xmlFile, err := os.Open(path)
	if err != nil {
		logger.Error.Println("Error during reading category metadata")
		logger.Error.Println(err)
	}
	defer xmlFile.Close()
	byteValue, _ := io.ReadAll(xmlFile)
	var catmetadata Catmetadata
	xml.Unmarshal(byteValue, &catmetadata)
	return catmetadata
}

// Descriptions of the category metadata.xml format

type Catmetadata struct {
	XMLName          xml.Name          `xml:"catmetadata"`
	Longdescriptions []Longdescription `xml:"longdescription"`
}

type Longdescription struct {
	XMLName xml.Name `xml:"longdescription"`
	Lang    string   `xml:"lang,attr"`
	Content string   `xml:",chardata"`
}
