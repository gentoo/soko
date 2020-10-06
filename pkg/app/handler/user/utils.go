package user

import (
	"html/template"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/models"
	"strings"
)

// renderAboutTemplate renders a specific about template
func renderUserTemplate(w http.ResponseWriter, r *http.Request, allProjects []*models.Project, pageName, page string) {
	templates := template.Must(
		template.Must(
			template.Must(
				template.Must(
					template.New(page).Funcs(template.FuncMap{
						"Contains":         contains,
						"ContainsInt":      containsInt,
						"CreateSlice":      createSlice,
						"GetPkgcheckClass": models.GetPkgcheckClass,
						"add": func(a, b int) int {
							return a + b
						},
					}).
						ParseGlob("web/templates/layout/*.tmpl")).
					ParseGlob("web/templates/user/preferences/*.tmpl")).
				ParseGlob("web/templates/user/userheader.tmpl")).
			ParseGlob("web/templates/user/" + page + ".tmpl"))

	templates.ExecuteTemplate(w, page+".tmpl", getPageData(pageName, allProjects, utils.GetUserPreferences(r)))
}

func contains(list []string, item string) bool {
	return strings.Contains("  "+strings.Join(list, "  ")+"  ", "  "+item+"  ")
}

func containsInt(list []int, item int) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func createSlice(n int) []int {
	slice := []int{}
	for i := 0; i <= n; i++ {
		slice = append(slice, i)
	}
	return slice
}

// getPageData returns the data used
// in all about templates
func getPageData(pageName string, allProjects []*models.Project, preferences models.UserPreferences) interface{} {
	return struct {
		Header          models.Header
		Application     models.Application
		PageName        string
		Projects        []*models.Project
		UserPreferences models.UserPreferences
	}{
		Header:          models.Header{Title: "User – ", Tab: "user"},
		Application:     utils.GetApplicationData(),
		PageName:        pageName,
		Projects:        allProjects,
		UserPreferences: preferences,
	}
}
