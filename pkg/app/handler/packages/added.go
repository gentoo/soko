// Used to show recently added versions

package packages

import (
	"net/http"
	"soko/pkg/app/handler/feeds"
)

// Added renders a template containing
// a list of 50 recently added versions
func Added(w http.ResponseWriter, r *http.Request) {
	addedVersions := getAddedVersions(50)
	RenderPackageTemplates("changedVersions", "changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Added", addedVersions), w)
}

func AddedFeed(w http.ResponseWriter, r *http.Request) {
	addedVersions := getAddedVersions(50)
	feeds.Changes(GetTextFuncMap(), CreateFeedData("Added", addedVersions), w)
}
