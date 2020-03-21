// Used to show recently keyworded versions

package packages

import "net/http"

// Keyworded renders a template containing
// a list of 50 recently keyworded versions
func Keyworded(w http.ResponseWriter, r *http.Request) {
	keywordedVersions := GetKeywordedVersions(50)
	RenderPackageTemplates("changedVersions", "changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Keyworded", keywordedVersions), w)
}
