// Used to show recently updated versions

package packages

import (
	"net/http"
	"soko/pkg/app/handler/feeds"
)

// Updated renders a template containing
// a list of 50 recently updated versions
func Updated(w http.ResponseWriter, r *http.Request) {
	updatedVersions := GetUpdatedVersions(50)
	RenderPackageTemplates("changedVersions", "changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Updated", updatedVersions), w)
}

func UpdatedFeed(w http.ResponseWriter, r *http.Request) {
	updatedVersions := GetUpdatedVersions(50)
	feeds.Changes(GetTextFuncMap(), CreateFeedData("Updated", updatedVersions), w)
}
