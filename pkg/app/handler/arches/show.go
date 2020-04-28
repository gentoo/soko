package arches

import (
	"net/http"
	"soko/pkg/app/handler/feeds"
	"soko/pkg/app/handler/packages"
	"strings"
)

// Show renders a template to show a list of recently changed version by arch
func Show(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path[len("/arches/"):], "/")
	if len(urlParts) == 2 {
		if urlParts[1] == "stable" {
			stabilizedVersions, err := getStabilizedVersionsForArch(urlParts[0], 50)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			renderPackageTemplates("changedVersions", packages.GetFuncMap(), createFeedData(urlParts[0], "Newly Stable", "stable", stabilizedVersions), w)
		} else if urlParts[1] == "stable.atom" {
			stabilizedVersions, err := getStabilizedVersionsForArch(urlParts[0], 50)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			feeds.Changes(packages.GetTextFuncMap(), createFeedData(urlParts[0], "Newly Stable", "stable", stabilizedVersions), w)
		} else if urlParts[1] == "keyworded" {
			keywordedVersions, err := getKeywordedVersionsForArch(urlParts[0], 50)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			renderPackageTemplates("changedVersions", packages.GetFuncMap(), createFeedData(urlParts[0], "Keyworded", "keyworded", keywordedVersions), w)
		} else if urlParts[1] == "keyworded.atom" {
			keywordedVersions, err := getKeywordedVersionsForArch(urlParts[0], 50)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			feeds.Changes(packages.GetTextFuncMap(), createFeedData(urlParts[0], "Keyworded", "keyworded", keywordedVersions), w)
		}
	}
}
