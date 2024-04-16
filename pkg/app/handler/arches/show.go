package arches

import (
	"net/http"
	"soko/pkg/app/handler/feeds"
	"soko/pkg/app/layout"
	"strings"
)

func ShowStable(w http.ResponseWriter, r *http.Request) {
	arch := r.PathValue("arch")
	stabilizedVersions, err := getStabilizedVersionsForArch(arch, 50)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	layout.Layout("Architectures", "arches", changedVersions(
		arch, "Newly Stable", "stable", stabilizedVersions,
	)).Render(r.Context(), w)
}

func ShowStableFeed(w http.ResponseWriter, r *http.Request) {
	arch := r.PathValue("arch")
	stabilizedVersions, err := getStabilizedVersionsForArch(arch, 250)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	feedTitle := "Stabilized packages in Gentoo on " + arch
	feedDescription := feedTitle
	feeds.Changes(feedTitle, feedDescription, stabilizedVersions, w)
}

func ShowKeyworded(w http.ResponseWriter, r *http.Request) {
	arch := r.PathValue("arch")
	keywordedVersions, err := getKeywordedVersionsForArch(arch, 50)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	layout.Layout("Architectures", "arches", changedVersions(
		arch, "Keyworded", "keyworded", keywordedVersions,
	)).Render(r.Context(), w)
}

func ShowKeywordedFeed(w http.ResponseWriter, r *http.Request) {
	arch := r.PathValue("arch")
	keywordedVersions, err := getKeywordedVersionsForArch(arch, 250)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	feedTitle := "Keyworded packages in Gentoo on " + arch
	feedDescription := feedTitle
	feeds.Changes(feedTitle, feedDescription, keywordedVersions, w)
}

func ShowLeafPackages(w http.ResponseWriter, r *http.Request) {
	arch := r.PathValue("arch")
	leafs, err := getLeafPackagesForArch(arch)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte(strings.Join(leafs, "\n")))
}
