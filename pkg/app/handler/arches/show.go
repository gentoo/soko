package arches

import (
	"net/http"
	"soko/pkg/app/handler/feeds"
	"soko/pkg/app/handler/packages"
	"soko/pkg/app/utils"
	"strings"
)

// Show renders a template to show a list of recently changed version by arch
func Show(w http.ResponseWriter, r *http.Request) {
	arch, subPage, _ := strings.Cut(r.URL.Path[len("/arches/"):], "/")
	switch subPage {
	case "stable":
		stabilizedVersions, err := getStabilizedVersionsForArch(arch, 50)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		renderPackageTemplates("changedVersions", packages.GetFuncMap(), createFeedData(arch, "Newly Stable", "stable", stabilizedVersions, utils.GetUserPreferences(r)), w)
	case "stable.atom":
		stabilizedVersions, err := getStabilizedVersionsForArch(arch, 250)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		feedTitle := "Stabilized packages in Gentoo on " + arch
		feedDescription := feedTitle
		feeds.Changes(feedTitle, feedDescription, stabilizedVersions, w)
	case "keyworded":
		keywordedVersions, err := getKeywordedVersionsForArch(arch, 50)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		renderPackageTemplates("changedVersions", packages.GetFuncMap(), createFeedData(arch, "Keyworded", "keyworded", keywordedVersions, utils.GetUserPreferences(r)), w)
	case "keyworded.atom":
		keywordedVersions, err := getKeywordedVersionsForArch(arch, 250)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		feedTitle := "Keyworded packages in Gentoo on " + arch
		feedDescription := feedTitle
		feeds.Changes(feedTitle, feedDescription, keywordedVersions, w)
	default:
		http.NotFound(w, r)
	}
}
