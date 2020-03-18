package arches

import (
	"net/http"
	"soko/pkg/app/handler/packages"
	"strings"
)

// Show renders a template to show a list of recently changed version by arch
func Show(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path[len("/arches/"):], "/")
	if(len(urlParts) == 2){
		if(urlParts[1] == "stable"){
			stabilizedVersions := getStabilizedVersionsForArch(urlParts[0], 50)
			renderPackageTemplates("changedVersions",packages.GetFuncMap(), createFeedData( urlParts[0], "Newly Stable", "stable", stabilizedVersions),w)
		} else if(urlParts[1] == "keyworded"){
			keywordedVersions := getKeywordedVersionsForArch(urlParts[0], 50)
			renderPackageTemplates("changedVersions",packages.GetFuncMap(), createFeedData( urlParts[0], "Keyworded", "keyworded", keywordedVersions),w)
		}
	}
}
