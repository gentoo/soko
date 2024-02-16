// miscellaneous utility functions used for packages

package packages

import (
	"crypto/md5"
	"encoding/hex"
	"html/template"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"sort"
	"strings"

	"github.com/go-pg/pg/v10"
)

// GetAddedPackages returns a list of recently added
// packages containing a given number of packages
func GetAddedPackages(n int) (addedPackages []*models.Package) {
	err := database.DBCon.Model(&addedPackages).
		Order("preceding_commits DESC").
		Limit(n).
		Relation("Versions").
		Relation("Versions.Commits").
		Select()
	if err != nil && err != pg.ErrNoRows {
		logger.Error.Println("Error during fetching added packages from database", err)
	}
	return
}

// GetAddedVersions returns a list of recently added
// versions containing a given number of versions
func GetAddedVersions(n int) (addedVersions []*models.Version) {
	addedPackages := GetAddedPackages(n)
	for _, addedPackage := range addedPackages {
		addedVersions = append(addedVersions, addedPackage.Versions...)
	}
	return
}

// GetUpdatedVersions returns a list of recently updated
// versions containing a given number of versions
func GetUpdatedVersions(n int) (updatedVersions []*models.Version) {
	var updates []models.Commit
	err := database.DBCon.Model(&updates).
		Order("preceding_commits DESC").
		Limit(n).
		Relation("ChangedVersions", func(q *pg.Query) (*pg.Query, error) {
			return q.Limit(10 * n), nil
		}).
		Relation("ChangedVersions.Commits", func(q *pg.Query) (*pg.Query, error) {
			return q.Order("preceding_commits DESC"), nil
		}).
		Select()
	if err != nil {
		return
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
	return
}

// GetStabilizedVersions returns a list of recently stabilized
// versions containing a given number of versions
func GetStabilizedVersions(n int) (stabilizedVersions []*models.Version) {
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("stabilized IS NOT NULL").
		Where("version.id IS NOT NULL").
		Limit(n).
		Select()
	if err != nil {
		return
	}

	stabilizedVersions = make([]*models.Version, len(updates))
	for i, update := range updates {
		update.Version.Commits = []*models.Commit{update.Commit}
		stabilizedVersions[i] = update.Version
	}
	return
}

// GetKeywordedVersions returns a list of recently keyworded
// versions containing a given number of versions
func GetKeywordedVersions(n int) (keywordedVersions []*models.Version) {
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("added IS NOT NULL").
		Where("version.id IS NOT NULL").
		Limit(n).
		Select()
	if err != nil {
		return
	}

	keywordedVersions = make([]*models.Version, len(updates))
	for i, update := range updates {
		update.Version.Commits = []*models.Commit{update.Commit}
		keywordedVersions[i] = update.Version
	}
	return
}

// RenderPackageTemplates renders the given templates using the given data
// One pattern can be used to specify templates
func renderPackageTemplate(page string, templatepattern string, funcMap template.FuncMap, data interface{}, w http.ResponseWriter) {
	templates := template.Must(
		template.Must(
			template.Must(
				template.New(page).
					Funcs(funcMap).
					ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/packages/components/*.tmpl")).
			ParseGlob("web/templates/packages/" + templatepattern + ".tmpl"))
	templates.ExecuteTemplate(w, page+".tmpl", data)
}

// getAtom returns the atom of the package from the given url
func getAtom(r *http.Request) string {
	atom := r.URL.Path[len("/packages/"):]
	atom = strings.Replace(atom, "/changelog", "", 1)
	atom = strings.Replace(atom, ".html", "", 1)
	atom = strings.Replace(atom, ".json", "", 1)
	return atom
}

// getSearchData returns the data used in search templates
func getSearchData(packages []models.Package, search string) interface{} {
	return struct {
		Header      models.Header
		Search      string
		Packages    []models.Package
		Application models.Application
	}{
		Header:      models.Header{Title: search + " – ", Tab: "packages"},
		Search:      search,
		Packages:    packages,
		Application: utils.GetApplicationData(),
	}
}

// GetFuncMap returns the FuncMap used in templates
func GetFuncMap() template.FuncMap {
	return template.FuncMap{
		"contains":           strings.Contains,
		"listContains":       listContains,
		"replaceall":         strings.ReplaceAll,
		"gravatar":           gravatar,
		"mkSlice":            mkSlice,
		"getReverse":         getReverse,
		"tolower":            strings.ToLower,
		"formatRestricts":    utils.FormatRestricts,
		"RemoteIdLink":       remoteIdLink,
		"isMasked":           isMasked,
		"getMask":            getMask,
		"noescape":           func(str string) template.HTML { return template.HTML(str) },
		"showRemovalNotice":  showRemovalNotice,
		"isDeprecated":       isDeprecated,
		"getDeprecation":     getDeprecation,
		"bugCategoriesCount": bugCategoriesCount,
	}
}

func listContains(list []string, value string) bool {
	for _, item := range list {
		if value == item {
			return true
		}
	}
	return false
}

// gravatar creates a link to the gravatar
// based on the email
func gravatar(email string) string {
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
func getReverse(index int, versions []*models.Version) *models.Version {
	return versions[len(versions)-1-index]
}

// getParameterValue returns the value of a given parameter
func getParameterValue(parameterName string, r *http.Request) string {
	results, ok := r.URL.Query()[parameterName]
	if !ok {
		return ""
	}
	if len(results) == 0 {
		return ""
	}
	return results[0]
}

// getPackageUseflags retrieves all local USE flags, global USE
// flags and use expands for a given package
func getPackageUseflags(gpackage *models.Package) ([]models.Useflag, []models.Useflag, map[string][]models.Useflag) {
	var localUseflags, allGlobalUseflags, filteredGlobalUseflags []models.Useflag
	useExpands := make(map[string][]models.Useflag)

	if len(gpackage.Versions) == 0 {
		return localUseflags, allGlobalUseflags, useExpands
	}

	rawUseFlags := make([]string, len(gpackage.Versions[0].Useflags))
	for i, rawUseflag := range gpackage.Versions[0].Useflags {
		rawUseFlags[i] = strings.Replace(rawUseflag, "+", "", 1)
	}

	if len(rawUseFlags) == 0 {
		return localUseflags, allGlobalUseflags, useExpands
	}

	var tmp_useflags []models.Useflag
	err := database.DBCon.Model(&tmp_useflags).
		Where("name in (?)", pg.In(rawUseFlags)).
		Order("name ASC").
		Select()
	if err != nil && err != pg.ErrNoRows {
		logger.Error.Println("Error during fetching added packages from database", err)
		return localUseflags, allGlobalUseflags, useExpands
	}

	for _, useflag := range tmp_useflags {
		if useflag.Scope == "global" {
			allGlobalUseflags = append(allGlobalUseflags, useflag)
		} else if useflag.Scope == "local" {
			if useflag.Package == gpackage.Atom {
				localUseflags = append(localUseflags, useflag)
			}
		} else {
			useExpands[useflag.UseExpand] = append(useExpands[useflag.UseExpand], useflag)
		}
	}

	// Only add global useflags that are not present in the local useflags
	for _, useflag := range allGlobalUseflags {
		if !containsUseflag(useflag, localUseflags) {
			filteredGlobalUseflags = append(filteredGlobalUseflags, useflag)
		}
	}

	return localUseflags, filteredGlobalUseflags, useExpands
}

// createPackageData creates the data used in the show package template
func createPackageData(
	pageName string,
	gpackage *models.Package,
	localUseflags, globalUseflags []models.Useflag, useExpands map[string][]models.Useflag,
	userPreferences models.UserPreferences,
	securityBugs, nonSecurityBugs int,
) interface{} {
	return struct {
		PageName             string
		Header               models.Header
		Package              models.Package
		Versions             []*models.Version
		Masks                []models.Mask
		LocalUseflags        []models.Useflag
		GlobalUseflags       []models.Useflag
		UseExpands           map[string][]models.Useflag
		Application          models.Application
		UserPreferences      models.UserPreferences
		SecurityBugsCount    int
		NonSecurityBugsCount int
	}{
		PageName:             pageName,
		Header:               models.Header{Title: gpackage.Atom + " – ", Tab: "packages"},
		Package:              *gpackage,
		Versions:             gpackage.Versions,
		LocalUseflags:        localUseflags,
		GlobalUseflags:       globalUseflags,
		UseExpands:           useExpands,
		Masks:                nil,
		Application:          utils.GetApplicationData(),
		UserPreferences:      userPreferences,
		SecurityBugsCount:    securityBugs,
		NonSecurityBugsCount: nonSecurityBugs,
	}
}

// CreateFeedData creates the data used in changedVersions template
func CreateFeedData(name string, versions []*models.Version) interface{} {
	return struct {
		Header      models.Header
		Name        string
		Versions    []*models.Version
		Application models.Application
	}{
		Header:      models.Header{Title: "Packages – ", Tab: "packages"},
		Name:        name,
		Versions:    versions,
		Application: utils.GetApplicationData(),
	}
}

// remoteIdLink returns a link to the homepage of a given remote id
func remoteIdLink(remoteId models.RemoteId) string {
	switch remoteId.Type {
	case "bitbucket":
		return "https://bitbucket.org/" + remoteId.Id
	case "codeberg":
		return "https://codeberg.org/" + remoteId.Id
	case "cpan":
		return "https://metacpan.org/dist/" + remoteId.Id
	case "cpan-module":
		return "https://metacpan.org/pod/" + remoteId.Id
	case "cran":
		return "https://cran.r-project.org/web/packages/" + remoteId.Id + "/"
	case "ctan":
		return "https://ctan.org/pkg/" + remoteId.Id
	case "freedesktop-gitlab":
		return "https://gitlab.freedesktop.org/" + remoteId.Id + ".git/"
	case "gentoo":
		return "https://gitweb.gentoo.org/" + remoteId.Id + ".git/"
	case "github":
		return "https://github.com/" + remoteId.Id
	case "gitlab":
		return "https://gitlab.com/" + remoteId.Id
	case "gnome-gitlab":
		return "https://gitlab.gnome.org/" + remoteId.Id + ".git/"
	case "google-code":
		return "https://code.google.com/archive/p/" + remoteId.Id + "/"
	case "hackage":
		return "https://hackage.haskell.org/package/" + remoteId.Id
	case "heptapod":
		return "https://foss.heptapod.net/" + remoteId.Id
	case "kde-invent":
		return "https://invent.kde.org/" + remoteId.Id
	case "launchpad":
		return "https://launchpad.net/" + remoteId.Id
	case "osdn":
		return "https://osdn.net/projects/" + remoteId.Id + "/"
	case "pear":
		return "https://pear.php.net/package/" + remoteId.Id
	case "pecl":
		return "https://pecl.php.net/package/" + remoteId.Id
	case "pypi":
		return "https://pypi.org/project/" + remoteId.Id + "/"
	case "rubygems":
		return "https://rubygems.org/gems/" + remoteId.Id + "/"
	case "savannah":
		return "https://savannah.gnu.org/projects/" + remoteId.Id
	case "savannah-nongnu":
		return "https://savannah.nongnu.org/projects/" + remoteId.Id
	case "sourceforge":
		return "https://sourceforge.net/projects/" + remoteId.Id + "/"
	case "sourcehut":
		return "https://sr.ht/" + remoteId.Id + "/"
	case "vim":
		return "https://vim.org/scripts/script.php?script_id=" + remoteId.Id
	default:
		return ""
	}
}

// isMasked returns true if any version is masked
func isMasked(versions []*models.Version) bool {
	for _, version := range versions {
		if len(version.Masks) > 0 {
			return true
		}
	}
	return false
}

// getMask returns the mask entry of the first version that is masked
func getMask(versions []*models.Version) *models.Mask {
	for _, version := range versions {
		if len(version.Masks) > 0 {
			return version.Masks[0]
		}
	}
	return nil
}

// showRemovalNotice if all versions of the package are masked
func showRemovalNotice(versions []*models.Version) bool {
	for _, version := range versions {
		if len(version.Masks) > 0 && version.Masks[0].Versions == version.Atom {
			return true
		}
	}
	return false
}

// isDeprecated returns true if any version is deprecated
func isDeprecated(versions []*models.Version) bool {
	for _, version := range versions {
		if len(version.Deprecates) > 0 {
			return true
		}
	}
	return false
}

// getDeprecation returns the deprecation entry of the first version that is deprecated
func getDeprecation(versions []*models.Version) *models.DeprecatedPackage {
	for _, version := range versions {
		if len(version.Deprecates) > 0 {
			return version.Deprecates[0]
		}
	}
	return nil
}

// sort the versions in descending order
func sortVersionsDesc(versions []*models.Version) {
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].GreaterThan(*versions[j])
	})
}

// containsUseflag returns true if the given list of useflags contains the
// given userflag. Otherwise false will be returned.
func containsUseflag(useflag models.Useflag, useflags []models.Useflag) bool {
	for _, flag := range useflags {
		if useflag.Name == flag.Name {
			return true
		}
	}
	return false
}

func bugCategoriesCount(bugs []*models.Bug) interface{} {
	var result struct {
		General, Stabilization, Keywording int
	}
	for _, bug := range bugs {
		switch bug.Component {
		case "Current packages":
			result.General++
		case "Stabilization":
			result.Stabilization++
		case "Keywording":
			result.Keywording++
		}
	}
	return result
}
