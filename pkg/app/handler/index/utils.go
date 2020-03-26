// miscellaneous utility functions used for the landing page of the application

package index

import (
	"github.com/go-pg/pg/v9/orm"
	"html/template"
	"net/http"
	"soko/pkg/app/handler/packages"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	"strconv"
	"strings"
)

// getAddedPackages returns a list of a
// given number of recently added Versions
func getAddedPackages(n int) []models.Package {
	var addedPackages []models.Package
	err := database.DBCon.Model(&addedPackages).
		Order("preceding_commits DESC").
		Limit(n).
		Relation("Versions").
		Select()
	if err != nil {
		return addedPackages
	}
	return addedPackages
}

// getUpdatedVersions returns a list of a
// given number of recently updated Versions
func getUpdatedVersions(n int) []*models.Version {
	var updatedVersions []*models.Version
	var updates []models.Commit
	err := database.DBCon.Model(&updates).
		Order("preceding_commits DESC").
		Limit(3*n).
		Relation("ChangedVersions", func(q *orm.Query) (*orm.Query, error) {
			return q.Limit(n), nil
		}).
		Relation("ChangedVersions.Commits", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("preceding_commits DESC").Limit(1), nil
		}).
		Select()
	if err != nil {
		return updatedVersions
	}
	for _, commit := range updates {
		updatedVersions = append(updatedVersions, commit.ChangedVersions...)
	}
	if len(updatedVersions) > n {
		updatedVersions = updatedVersions[:n]
	}
	return updatedVersions
}

// createPageData creates the data used in the template of the landing page
func createPageData(packagecount int, addedPackages []models.Package, updatedVersions []*models.Version) interface{} {
	return struct {
		Page            string
		PackageCount    string
		AddedPackages   []models.Package
		UpdatedPackages []*models.Version
		Application     models.Application
	}{
		Page:            "home",
		Application:     utils.GetApplicationData(),
		PackageCount:    formatPackageCount(packagecount),
		AddedPackages:   addedPackages,
		UpdatedPackages: updatedVersions,
	}
}

// renderIndexTemplate renders all templates used for the landing page
func renderIndexTemplate(w http.ResponseWriter, pageData interface{}) {
	templates := template.Must(
		template.Must(
			template.Must(
				template.New("Show").
					Funcs(getFuncMap()).
					ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/packages/changedVersionRow.tmpl")).
			ParseGlob("web/templates/index/*.tmpl"))

	templates.ExecuteTemplate(w, "show.tmpl", pageData)
}

// GetFuncMap returns the FuncMap used in templates
func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"contains":        strings.Contains,
		"mkSlice":         mkSlice,
		"formatRestricts": packages.FormatRestricts,
	}
}

// formatPackageCount returns the formatted number of
// packages containing a thousands comma
func formatPackageCount(packageCount int) string {
	packages := strconv.Itoa(packageCount)
	if len(string(packageCount)) == 6 {
		return packages[:3] + "," + packages[3:]
	} else if len(packages) == 5 {
		return packages[:2] + "," + packages[2:]
	} else if len(packages) == 4 {
		return packages[:1] + "," + packages[1:]
	} else {
		return packages
	}
}

// mkSlice creates a slice based on the given arguments
func mkSlice(args ...interface{}) []interface{} {
	return args
}
