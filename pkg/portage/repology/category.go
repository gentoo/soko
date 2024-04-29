package repology

import (
	"log/slog"

	"soko/pkg/database"
	"soko/pkg/models"
)

func UpdateCategoriesMetadata() {
	var categoriesInfoArr []*models.CategoryPackagesInformation
	err := database.DBCon.Model((*models.OutdatedPackages)(nil)).
		ColumnExpr("SPLIT_PART(atom, '/', 1) as name").
		ColumnExpr("COUNT(*) as outdated").
		GroupExpr("SPLIT_PART(atom, '/', 1)").
		Select(&categoriesInfoArr)
	if err != nil {
		slog.Error("Failed collecting outdated stats", slog.Any("err", err))
		return
	}
	categoriesInfo := make(map[string]*models.CategoryPackagesInformation, len(categoriesInfoArr))
	for _, categoryInfo := range categoriesInfoArr {
		if categoryInfo.Name != "" {
			categoriesInfo[categoryInfo.Name] = categoryInfo
		}
	}

	var categories []*models.CategoryPackagesInformation
	err = database.DBCon.Model(&categories).Column("name").Select()
	if err != nil {
		slog.Error("Failed fetching categories packages information", slog.Any("err", err))
		return
	} else if len(categories) > 0 {
		for _, category := range categories {
			if info, found := categoriesInfo[category.Name]; found {
				category.Outdated = info.Outdated
			} else {
				category.Outdated = 0
			}
			delete(categoriesInfo, category.Name)
		}
		_, err = database.DBCon.Model(&categories).Set("outdated = ?outdated").Update()
		if err != nil {
			slog.Error("Failed updating categories packages information", slog.Any("err", err))
		}
		categories = make([]*models.CategoryPackagesInformation, 0, len(categoriesInfo))
	}

	for _, catInfo := range categoriesInfo {
		categories = append(categories, catInfo)
	}
	if len(categories) > 0 {
		_, err = database.DBCon.Model(&categories).Insert()
		if err != nil {
			slog.Error("Failed inserting categories packages information", slog.Any("err", err))
		}
	}
}
