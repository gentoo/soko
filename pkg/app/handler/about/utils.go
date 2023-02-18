// miscellaneous utility functions used for the about pages

package about

import (
	"html/template"
	"net/http"
	"runtime/debug"
	"soko/pkg/app/utils"
	"soko/pkg/models"
)

// renderAboutTemplate renders a specific about template
func renderAboutTemplate(w http.ResponseWriter, r *http.Request, page string) {
	templates := template.Must(
		template.Must(
			template.New(page).
				ParseGlob("web/templates/layout/*.tmpl")).
			ParseGlob("web/templates/about/" + page + ".tmpl"))

	templates.ExecuteTemplate(w, page+".tmpl", getPageData())
}

func getCommitId() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:8]
			}
		}
	}
	return ""
}

// getPageData returns the data used
// in all about templates
func getPageData() interface{} {
	return struct {
		Header      models.Header
		Application models.Application
		CommitId    string
	}{
		Header:      models.Header{Title: "About – ", Tab: "about"},
		Application: utils.GetApplicationData(),
		CommitId:    getCommitId(),
	}
}
