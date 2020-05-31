package feeds

import (
	"fmt"
	"github.com/gorilla/feeds"
	"net/http"
	"soko/pkg/models"
	"time"
)

// Show renders a template to show a given package
func Packages(query string, gpackages []models.Package, w http.ResponseWriter) {
	feed := &feeds.Feed{
		Title:       "Gentoo Packages for search query: " + query,
		Description: "Gentoo Packages for search query: " + query,
		Author:      &feeds.Author{Name: "Gentoo Packages Database"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: "https://packages.gentoo.org/"},
	}
	addPackageFeedItems(feed, gpackages)
	feed.WriteAtom(w)
}

// addPackageFeedItems is a helper to add items to a feed; the Package feed is using []models.Package as the entity.
func addPackageFeedItems(f *feeds.Feed, gpackages []models.Package) {
	for _, gpackage := range gpackages {
		item := &feeds.Item{
			Title:  gpackage.Atom,
			Link: &feeds.Link{Href: fmt.Sprintf("https://packages.gentoo.org/package/%s", gpackage.Atom)},
			Description: gpackage.Longdescription,
			Author:      &feeds.Author{Name: "unknown"},
			Created:  time.Now(),
		}
		f.Add(item)
	}
}