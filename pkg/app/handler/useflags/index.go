// Used to display the landing page of the USE flag section

package useflags

import (
	"net/http"
	"soko/pkg/app/utils"
)

// Index renders a template to show the index page of the USE flags
// section containing a bubble chart of popular USE flags
func Default(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetUserPreferences(r)
	if userPreferences.Useflags.Layout == "bubble" {
		http.Redirect(w, r, "/useflags/popular", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/useflags/search", http.StatusSeeOther)
	}
}
