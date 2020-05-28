// Used to show recently added versions

package packages

import (
	"github.com/gorilla/feeds"
	"net/http"
	"time"
)

// Added renders a template containing a list of 50 recently added versions.
func Added(w http.ResponseWriter, r *http.Request) {
	addedVersions := getAddedVersions(50)
	RenderPackageTemplates("changedVersions", "changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Added", addedVersions), w)
}

func AddedFeed(w http.ResponseWriter, r *http.Request) {
	addedVersions := getAddedVersions(250)
	feed := &feeds.Feed{
		Title:       "Added packages in Gentoo.",
		Description: "Added packages in Gentoo.",
		Author:      &feeds.Author{Name: "Gentoo Packages Database"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: "https://packages.gentoo.org"},
	}
	addFeedItems(feed, addedVersions)
	feed.WriteAtom(w)
}
