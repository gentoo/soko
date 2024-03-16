// Contains functions to import categories into the database

package repository

import (
	"encoding/xml"
	"log/slog"
	"os"
	"regexp"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
)

// isCategory checks whether the path points to a category
// descriptions that is an metadata.xml file
func isCategory(path string) bool {
	isCategory, _ := regexp.MatchString(`^[^/]*\/metadata\.xml$`, path)
	return isCategory
}

// UpdateCategories updates the categories in the database for each
// given path that points to a category description
func UpdateCategories(paths []string) {
	deleted := map[string]*models.Category{}
	modified := map[string]*models.Category{}

	for _, path := range paths {
		splittedLine := strings.Split(path, "\t")

		if len(splittedLine) != 2 {
			if len(splittedLine) == 1 && isCategory(path) {
				if cat := updateModifiedCategory(path); cat != nil {
					modified[cat.Name] = cat
				}
			}
			continue
		}

		status := splittedLine[0]
		changedFile := splittedLine[1]

		if !isCategory(changedFile) {
			continue
		} else if status == "D" {
			cat := updateDeletedCategory(changedFile)
			deleted[cat.Name] = cat
		} else if status == "A" || status == "M" {
			if cat := updateModifiedCategory(changedFile); cat != nil {
				modified[cat.Name] = cat
			}
		}
	}

	if len(deleted) > 0 {
		rows := make([]*models.Category, 0, len(deleted))
		for _, row := range deleted {
			rows = append(rows, row)
		}
		res, err := database.DBCon.Model(&rows).Delete()
		if err != nil {
			slog.Error("Failed deleting categories", slog.Any("err", err))
		} else {
			slog.Info("Deleted categories", slog.Int("rows", res.RowsAffected()))
		}
	}

	if len(modified) > 0 {
		rows := make([]*models.Category, 0, len(modified))
		for _, row := range modified {
			rows = append(rows, row)
		}
		res, err := database.DBCon.Model(&rows).OnConflict("(name) DO UPDATE").Insert()
		if err != nil {
			slog.Error("Failed updating categories", slog.Any("err", err))
		} else {
			slog.Info("Updated categories", slog.Int("rows", res.RowsAffected()))
		}
	}
}

// updateDeletedCategory deletes a category from the database
func updateDeletedCategory(changedFile string) *models.Category {
	name, _, _ := strings.Cut(changedFile, "/")
	return &models.Category{Name: name}
}

// updateModifiedCategory adds a category to the database or
// updates it. To do so, it parses the metadata from metadata.xml
func updateModifiedCategory(changedFile string) *models.Category {
	name, _, _ := strings.Cut(changedFile, "/")

	xmlFile, err := os.Open(config.PortDir() + "/" + changedFile)
	if err != nil {
		slog.Error("Failed reading category metadata", slog.String("category", changedFile), slog.Any("err", err))
		return nil
	}
	defer xmlFile.Close()

	var catMetadata CatMetadata
	err = xml.NewDecoder(xmlFile).Decode(&catMetadata)
	if err != nil {
		slog.Error("Error decoding category", slog.String("category", changedFile), slog.Any("err", err))
		return nil
	}

	var description string
	for _, longDescription := range catMetadata.LongDescriptions {
		if longDescription.Lang == "en" || longDescription.Lang == "" {
			description = strings.TrimSpace(longDescription.Content)
		}
	}

	return &models.Category{
		Name:        name,
		Description: description,
	}
}

// Descriptions of the category metadata.xml format

type CatMetadata struct {
	XMLName          xml.Name          `xml:"catmetadata"`
	LongDescriptions []LongDescription `xml:"longdescription"`
}

type LongDescription struct {
	XMLName xml.Name `xml:"longdescription"`
	Lang    string   `xml:"lang,attr"`
	Content string   `xml:",chardata"`
}
