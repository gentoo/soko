// SPDX-License-Identifier: GPL-2.0-only

// miscellaneous utility functions used for arches

package arches

import (
	"soko/pkg/database"
	"soko/pkg/models"
)

// getStabilizedVersionsForArch returns the given number of recently
// stabilized versions of a specific arch
func getStabilizedVersionsForArch(arch string, n int) ([]*models.Version, error) {
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("stabilized::jsonb @> ?", "\""+arch+"\"").
		Where("version.id IS NOT NULL").
		Limit(n).
		Select()
	if err != nil {
		return nil, err
	}

	stabilizedVersions := make([]*models.Version, len(updates))
	for i, update := range updates {
		update.Version.Commits = []*models.Commit{update.Commit}
		stabilizedVersions[i] = update.Version
	}
	return stabilizedVersions, err
}

// getKeywordedVersionsForArch returns the given number of recently
// keyworded versions of a specific arch
func getKeywordedVersionsForArch(arch string, n int) ([]*models.Version, error) {
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("added::jsonb @> ?", "\""+arch+"\"").
		Where("version.id IS NOT NULL").
		Limit(n).
		Select()
	if err != nil {
		return nil, err
	}

	keywordedVersions := make([]*models.Version, len(updates))
	for i, update := range updates {
		update.Version.Commits = []*models.Commit{update.Commit}
		keywordedVersions[i] = update.Version
	}
	return keywordedVersions, err
}

func getLeafPackagesForArch(arch string) ([]string, error) {
	var atoms []string
	atomsWithReverse := database.DBCon.Model((*models.ReverseDependency)(nil)).
		Join("JOIN versions").JoinOn("reverse_dependency.reverse_dependency_atom = versions.atom").
		Where("? = ANY(STRING_TO_ARRAY(keywords, ' '))", arch).
		WhereOr("? = ANY(STRING_TO_ARRAY(keywords, ' '))", "~"+arch).
		ColumnExpr("DISTINCT reverse_dependency.atom")
	err := database.DBCon.Model((*models.Version)(nil)).
		Where("((? = ANY(STRING_TO_ARRAY(keywords, ' '))) OR (? = ANY(STRING_TO_ARRAY(keywords, ' ')))) AND (atom NOT IN (?))",
			arch, "~"+arch, atomsWithReverse).
		Order("atom").
		ColumnExpr("DISTINCT atom").
		Select(&atoms)
	return atoms, err
}
