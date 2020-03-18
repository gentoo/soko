// Used to show recently added versions

package packages

import "net/http"

// Added renders a template containing
// a list of 50 recently added versions
func Added(w http.ResponseWriter, r *http.Request) {
	addedVersions := getAddedVersions(50)
	RenderPackageTemplates("changedVersions","changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Added", addedVersions),w)
}
