// Used to show recently added versions

package packages

import (
	"fmt"
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
	for _, added := range addedVersions {
		cpv := fmt.Sprintf("%s-%s", added.Atom, added.Version)
		item := &feeds.Item{
			Title:       cpv
			Link:        &feeds.Link{Href: fmt.Sprintf("https://packages.gentoo.org/package/%s", added.Atom)},
			Description: added.Description,
			Author:      &feeds.Author{Name: "Unknown"},
			Created:     time.Now(),
		}
		if len(added.Commits) > 0 {
			lastCommit := added.Commits[0]
			item.Author = &feeds.Author{Name: lastCommit.CommitterName}
			item.Created = lastCommit.CommitterDate
			item.Content = fmt.Sprintf("%s is now available in Gentoo on these architectures: %s. See <a href='https://gitweb.gentoo.org/repo/gentoo.git/commit/?id=%s'>Gitweb</a>",
				cpv, added.Keywords, lastCommit.Id)
		}
		feed.Add(item)
	}
	feed.WriteAtom(w)
}
