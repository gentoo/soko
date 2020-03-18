// miscellaneous utility functions used for arches

package arches

import (
	"html/template"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
)

// getPageData creates the data used in all
// templates used in the arches section
func getPageData() interface{}{
	return struct {
		Page            string
		Application     models.Application
	}{
		Page:           "arches",
		Application:    utils.GetApplicationData(),
	}
}

// getStabilizedVersionsForArch returns the given number of recently
// stabilized versions of a specific arch
func getStabilizedVersionsForArch(arch string, n int) []*models.Version {
	var stabilizedVersions []*models.Version
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("stabilized::jsonb @> ?", "\"" + arch + "\"").
		Limit(n).
		Select()
	if err != nil {
		panic(err)
	}

	for _, update := range updates{
		if(update.Version != nil){
			update.Version.Commits = []*models.Commit{update.Commit}
			stabilizedVersions = append(stabilizedVersions, update.Version)
		}
	}

	return stabilizedVersions
}

// getKeywordedVersionsForArch returns the given number of recently
// keyworded versions of a specific arch
func getKeywordedVersionsForArch(arch string, n int) []*models.Version {
	var stabilizedVersions []*models.Version
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("added::jsonb @> ?", "\"" + arch + "\"").
		Limit(n).
		Select()
	if err != nil {
		panic(err)
	}

	for _, update := range updates{
		if(update.Version != nil){
			update.Version.Commits = []*models.Commit{update.Commit}
			stabilizedVersions = append(stabilizedVersions, update.Version)
		}
	}

	return stabilizedVersions
}

// RenderPackageTemplates renders the arches templates using the given data
func renderPackageTemplates(page string, funcMap template.FuncMap, data interface{}, w http.ResponseWriter){

	templates := template.Must(
					template.Must(
						template.Must(
							template.New(page).
								Funcs(funcMap).
								ParseGlob("web/templates/layout/*.tmpl")).
								ParseGlob("web/templates/packages/changedVersionRow.tmpl")).
								ParseGlob("web/templates/arches/changedVersions.tmpl"))

	templates.ExecuteTemplate(w, page+".tmpl", data)
}

// CreateFeedData creates the data used in changedVersions template
func createFeedData(arch string, name string, feedtype string, versions []*models.Version) interface{}{
	return struct {
		Page                  string
		Arch                  string
		Name                  string
		FeedName              string
		Versions            []*models.Version
		Application           models.Application
	}{
		Page:                 "arches",
		Arch:                 arch,
		Name:                 name,
		FeedName:             feedtype,
		Versions:             versions,
		Application:          utils.GetApplicationData(),
	}
}
