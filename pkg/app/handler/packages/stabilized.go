// Used to show recently stabilized versions

package packages

import (
	"net/http"
	"soko/pkg/app/handler/feeds"
)

// Stabilized renders a template containing
// a list of 50 recently stabilized versions
func Stabilized(w http.ResponseWriter, r *http.Request) {
	stabilizedVersions := GetStabilizedVersions(50)
	RenderPackageTemplates("changedVersions", "changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Stabilized", stabilizedVersions), w)
}

func StabilizedFeed(w http.ResponseWriter, r *http.Request) {
	stabilizedVersions := GetStabilizedVersions(1000)
	feeds.Changes("Stabilized packages in Gentoo.", "Stabilized packages in Gentoo.", stabilizedVersions, w)
}
