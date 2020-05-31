package feeds

import (
	"fmt"
	"github.com/gorilla/feeds"
	"net/http"
	"soko/pkg/models"
	"time"
)

// Show renders a template to show a given package
func Changes(title string, description string, changedVersions []*models.Version, w http.ResponseWriter) {
	feed := &feeds.Feed{
		Title:       title,
		Description: description,
		Author:      &feeds.Author{Name: "Gentoo Packages Database"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: "https://packages.gentoo.org/"},
	}
	addFeedItems(feed, changedVersions)
	feed.WriteAtom(w)

}


// addFeedItems is a helper to add items to a feed; most of the feeds use []*models.Version as the entity.
func addFeedItems(f *feeds.Feed, versions []*models.Version) {
	for _, version := range versions {
		cpv := fmt.Sprintf("%s-%s", version.Atom, version.Version)
		item := &feeds.Item{
			Title:  cpv,
			Link: &feeds.Link{Href: fmt.Sprintf("https://packages.gentoo.org/package/%s", version.Atom)},
			Description: version.Description,
			Author:      &feeds.Author{Name: "unknown"},
			Created:  time.Now(),
		}
		if len(version.Commits) > 0 {
			lastCommit := version.Commits[0]
			item.Author = &feeds.Author{Name: lastCommit.CommitterName}
			item.Created = lastCommit.CommitterDate
			item.Content = fmt.Sprintf("%s is now available in Gentoo on these architectures: %s. See <a href='https://gitweb.gentoo.org/repo/gentoo.git/commit/?id=%s'>Gitweb</a>",
				cpv, version.Keywords, lastCommit.Id)
		}
		f.Add(item)
	}
}