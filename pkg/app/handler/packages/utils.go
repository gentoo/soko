// miscellaneous utility functions used for packages

package packages

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/go-pg/pg/v9/orm"
	"html/template"
	"log"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	utils2 "soko/pkg/utils"
	"strings"
)

// getAddedPackages returns a list of recently added
// packages containing a given number of packages
func getAddedPackages(n int) []*models.Package {
	var addedPackages []*models.Package
	err := database.DBCon.Model(&addedPackages).
		Order("preceding_commits DESC").
		Limit(n).
		Relation("Versions").
		Relation("Versions.Commits").
		Select()
	if err != nil {
		panic(err)
	}
	return addedPackages
}

// getAddedVersions returns a list of recently added
// versions containing a given number of versions
func getAddedVersions(n int) []*models.Version {
	addedPackages := getAddedPackages(n)
	var addedVersions []*models.Version
	for _, addedPackage := range addedPackages{
		addedVersions = append(addedVersions, addedPackage.Versions...)
	}
	return addedVersions
}

// GetUpdatedVersions returns a list of recently updated
// versions containing a given number of versions
func GetUpdatedVersions(n int) []*models.Version {
	var updatedVersions []*models.Version
	var updates []models.Commit
	err := database.DBCon.Model(&updates).
		Order("preceding_commits DESC").
		Limit(n).
		Relation("ChangedVersions").
		Relation("ChangedVersions.Commits", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("preceding_commits DESC"), nil
		}).
		Select()
	if err != nil {
		panic(err)
	}
	for _, commit := range updates{
		for _, changedVersion := range commit.ChangedVersions{
			changedVersion.Commits = changedVersion.Commits[:1]
		}
		updatedVersions = append(updatedVersions, commit.ChangedVersions...)
	}
	if(len(updatedVersions) > n){
		updatedVersions = updatedVersions[:10]
	}
	return updatedVersions
}

// GetStabilizedVersions returns a list of recently stabilized
// versions containing a given number of versions
func GetStabilizedVersions(n int) []*models.Version {
	var stabilizedVersions []*models.Version
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("stabilized IS NOT NULL").
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

// GetKeywordedVersions returns a list of recently keyworded
// versions containing a given number of versions
func GetKeywordedVersions(n int) []*models.Version {
	var stabilizedVersions []*models.Version
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("added IS NOT NULL").
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

// RenderPackageTemplates renders the given templates using the given data
// One pattern can be used to specify templates
func renderPackageTemplate(page string, templatepattern string, funcMap template.FuncMap, data interface{}, w http.ResponseWriter){
	templates := template.Must(
					template.Must(
						template.New(page).
							Funcs(funcMap).
							ParseGlob("web/templates/layout/*.tmpl")).
							ParseGlob("web/templates/packages/" + templatepattern + ".tmpl"))
	templates.ExecuteTemplate(w, page+".tmpl", data)
}

// RenderPackageTemplates renders the given templates using the given data
// Two patterns can be used to specify templates
func RenderPackageTemplates(page string, templatepattern1 string, templatepattern2 string, funcMap template.FuncMap, data interface{}, w http.ResponseWriter){
	templates := template.Must(
		template.Must(
			template.Must(
				template.New(page).
					Funcs(funcMap).
					ParseGlob("web/templates/layout/*.tmpl")).
					ParseGlob("web/templates/packages/" + templatepattern1 + ".tmpl")).
					ParseGlob("web/templates/packages/" + templatepattern2 + ".tmpl"))
	templates.ExecuteTemplate(w, page+".tmpl", data)
}

// getAtom returns the atom of the package from the given url
func getAtom(r *http.Request) string {
	atom := r.URL.Path[len("/packages/"):]
	return strings.Replace(atom, "/changelog.html", "", 1)
}

// getSearchData returns the data used in search templates
func getSearchData(packages []models.Package, search string) interface{}{
	return struct{
		Page           string
		Search         string
		Packages     []models.Package
		Application    models.Application
	}{
		Page:          "packages",
		Search:        search,
		Packages:      packages,
		Application:   utils.GetApplicationData(),
	}
}

// getChangelogData returns the data used in changelog templates
func getChangelogData(commits []*models.Commit, atom string) interface{}{
	return struct {
		Commits  []*models.Commit
		Atom       string
	}{
		Commits:     commits,
		Atom:        atom,
	}
}

// GetFuncMap returns the FuncMap used in templates
func GetFuncMap() template.FuncMap{
	return template.FuncMap{
		"contains":     strings.Contains,
		"replaceall":   strings.ReplaceAll,
		"gravatar":     gravatar,
		"mkSlice":      mkSlice,
		"getReverse":   getReverse,
		"tolower":		strings.ToLower,
		"formatRestricts": FormatRestricts,
	}
}

// gravatar creates a link to the gravatar
// based on the email
func gravatar(email string) string{
	hasher := md5.Sum([]byte(email))
	hash := hex.EncodeToString(hasher[:])
	return "https://www.gravatar.com/avatar/" + hash + "?s=13&amp;d=retro"
}

// mkSlice creates a slice of the given arguments
func mkSlice(args ...interface{}) []interface{} {
	return args
}

// getReverse returns the element of a slice in
// reverse direction based on the index
func getReverse(index int, versions []*models.Version) *models.Version{
	return versions[len(versions)-1-index]
}

// getParameterValue returns the value of a given parameter
func getParameterValue(parameterName string, r *http.Request) string {
	results, _ := r.URL.Query()[parameterName]
	param := results[0]
	return param
}

// getPackageUseflags retrieves all local USE flags, global USE
// flags and use expands for a given package
func getPackageUseflags(gpackage *models.Package) ([]models.Useflag, []models.Useflag, []models.Useflag) {
	var localUseflags, globalUseflags, useExpands []models.Useflag
	for _, useflag := range gpackage.Versions[len(gpackage.Versions)-1].Useflags{

		log.Println(useflag)
		var tmp_useflags []models.Useflag
		err := database.DBCon.Model(&tmp_useflags).
			Where("Name LIKE ?", "%" + strings.Replace(useflag, "+", "", 1)).
			Select()

		if err != nil {
			panic(err)
		}

		if(len(tmp_useflags) >= 1 && tmp_useflags[0].Scope == "global"){
			globalUseflags = append(globalUseflags, tmp_useflags[0])
		}else if(len(tmp_useflags) >= 1 && tmp_useflags[0].Scope == "local") {
			localUseflags = append(localUseflags, tmp_useflags[0])
		}else if(len(tmp_useflags) >= 1){
			useExpands = append(useExpands, tmp_useflags[0])
		}
	}
	return localUseflags, globalUseflags, useExpands
}


// createPackageData creates the data used in the show package template
func createPackageData(gpackage *models.Package, localUseflags []models.Useflag, globalUseflags []models.Useflag, useExpands []models.Useflag) interface{}{
	return struct {
		Page                  string
		Package               models.Package
		Versions            []*models.Version
		Masks               []models.Mask
		LocalUseflags       []models.Useflag
		GlobalUseflags      []models.Useflag
		UseExpands          []models.Useflag
		Application           models.Application
	}{
		Page:                 "packages",
		Package:              *gpackage,
		Versions:             gpackage.Versions,
		LocalUseflags:        localUseflags,
		GlobalUseflags:       globalUseflags,
		UseExpands:           useExpands,
		Masks:                nil,
		Application:          utils.GetApplicationData(),
	}
}

// CreateFeedData creates the data used in changedVersions template
func CreateFeedData(name string, versions []*models.Version) interface{}{
	return struct {
		Page                  string
		Name                  string
		Versions            []*models.Version
		Application           models.Application
	}{
		Page:                 "packages",
		Name:                 name,
		Versions:             versions,
		Application:          utils.GetApplicationData(),
	}
}

// FormatRestricts returns a string containing a comma separated
// list of capitalized first letters of the package restricts
func FormatRestricts(restricts []string) string {
	var result []string
	for _, restrict := range restricts{
		if restrict != "(" && restrict != ")" && !strings.HasSuffix(restrict, "?") {
			result = append(result, strings.ToUpper(string(restrict[0])))
		}
	}
	result = utils2.Deduplicate(result)
	return strings.Join(result, ", ")
}
