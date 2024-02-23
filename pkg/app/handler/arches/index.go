package arches

import (
	"net/http"
	"soko/pkg/app/utils"
)

// Index renders a template to show a the landing page containing links to all arches feeds
func Index(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetUserPreferences(r)
	http.Redirect(w, r, "/arches/"+userPreferences.Arches.DefaultArch+"/"+userPreferences.Arches.DefaultPage, http.StatusSeeOther)
}
