// Used to show recently stabilized versions

package packages

import "net/http"

// Stabilized renders a template containing
// a list of 50 recently stabilized versions
func Stabilized(w http.ResponseWriter, r *http.Request) {
	stabilizedVersions := GetStabilizedVersions(50)
	RenderPackageTemplates("changedVersions","changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Stabilized", stabilizedVersions),w)
}
