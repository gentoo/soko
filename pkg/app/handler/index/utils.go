// miscellaneous utility functions used for the landing page of the application

package index

import (
	b64 "encoding/base64"
	"net/http"
	"slices"
	"soko/pkg/database"
	"soko/pkg/models"
	"strconv"
	"strings"

	"github.com/go-pg/pg/v10"
)

type packageInfo struct {
	Name        string
	Category    string
	Description string
}

// getAddedPackages returns a list of a
// given number of recently added Versions
func getAddedPackages(n int) (packages []packageInfo) {
	descriptionQuery := database.DBCon.Model((*models.Version)(nil)).
		Column("description").
		Where("atom = package.atom").
		Limit(1)
	err := database.DBCon.Model((*models.Package)(nil)).
		Column("name", "category").
		ColumnExpr("(?) AS description", descriptionQuery).
		Order("preceding_commits DESC").
		Limit(n).
		Select(&packages)
	if err != nil {
		return nil
	}
	return
}

func getSearchHistoryPackages(r *http.Request) (packages []packageInfo) {
	cookie, err := r.Cookie("search_history")
	if err != nil {
		return nil
	}
	packagesList := getSearchHistoryFromCookie(cookie)

	descriptionQuery := database.DBCon.Model((*models.Version)(nil)).
		Column("description").
		Where("atom = package.atom").
		Limit(1)
	err = database.DBCon.Model((*models.Package)(nil)).
		Column("name", "category").
		ColumnExpr("(?) AS description", descriptionQuery).
		Where("atom in (?)", pg.In(packagesList)).
		Select(&packages)
	if err != nil {
		return nil
	}

	return getSortedSearchHistory(packagesList, packages)
}

func getSortedSearchHistory(sortedPackagesList []string, packagesList []packageInfo) (result []packageInfo) {
	for _, gpackage := range sortedPackagesList {
		for _, gpackageObject := range packagesList {
			if gpackageObject.Category+"/"+gpackageObject.Name == gpackage {
				result = append(result, gpackageObject)
			}
		}
	}
	slices.Reverse(result)
	return
}

func getSearchHistoryFromCookie(cookie *http.Cookie) (packagesList []string) {
	cookieValue, err := b64.StdEncoding.DecodeString(cookie.Value)
	if err == nil {
		packagesList = strings.Split(string(cookieValue), ",")
		if len(packagesList) > 10 {
			packagesList = packagesList[len(packagesList)-10:]
		}
	}
	return
}

// getUpdatedVersions returns a list of a
// given number of recently updated Versions
func getUpdatedVersions(n int) []*models.Version {
	var updatedVersions []*models.Version
	var updates []models.Commit
	err := database.DBCon.Model(&updates).
		Order("preceding_commits DESC").
		Limit(3*n).
		Relation("ChangedVersions", func(q *pg.Query) (*pg.Query, error) {
			return q.Limit(30 * n), nil
		}).
		Relation("ChangedVersions.Commits", func(q *pg.Query) (*pg.Query, error) {
			return q.Order("preceding_commits DESC"), nil
		}).
		Select()
	if err != nil {
		return updatedVersions
	}
	for _, commit := range updates {
		for _, changedVersion := range commit.ChangedVersions {
			changedVersion.Commits = changedVersion.Commits[:1]
		}
		updatedVersions = append(updatedVersions, commit.ChangedVersions...)
	}
	if len(updatedVersions) > n {
		updatedVersions = updatedVersions[:n]
	}
	return updatedVersions
}

// formatPackageCount returns the formatted number of
// packages containing a thousands comma
func formatPackageCount(packageCount int) string {
	packages := strconv.Itoa(packageCount)
	if len(packages) == 6 {
		return packages[:3] + "," + packages[3:]
	} else if len(packages) == 5 {
		return packages[:2] + "," + packages[2:]
	} else if len(packages) == 4 {
		return packages[:1] + "," + packages[1:]
	} else {
		return packages
	}
}
