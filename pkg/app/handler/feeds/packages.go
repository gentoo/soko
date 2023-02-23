package feeds

import (
	"fmt"
	"net/http"
	"soko/pkg/models"
	"time"

	"github.com/gorilla/feeds"
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
			Title:       gpackage.Atom,
			Link:        &feeds.Link{Href: "https://packages.gentoo.org/package/" + gpackage.Atom},
			Description: gpackage.Longdescription,
			Author:      &feeds.Author{Name: "unknown"},
			Created:     time.Now(),
		}
		f.Add(item)
	}
}

// AddedPackages creates a feed for added packages
func AddedPackages(title string, description string, addedPackages []*models.Package, w http.ResponseWriter) {
	feed := &feeds.Feed{
		Title:       title,
		Description: description,
		Author:      &feeds.Author{Name: "Gentoo Packages Database"},
		Created:     time.Now(),
		Link:        &feeds.Link{Href: "https://packages.gentoo.org/"},
	}
	addAddedPackageFeedItems(feed, addedPackages)
	feed.WriteAtom(w)
}

// addAddedPackageFeedItems is a helper to add items to the added packages feed;
// it's using models.Package instead of models.Version
func addAddedPackageFeedItems(f *feeds.Feed, packages []*models.Package) {
	for _, gpackage := range packages {
		item := &feeds.Item{
			Title:       gpackage.Atom,
			Link:        &feeds.Link{Href: "https://packages.gentoo.org/package/" + gpackage.Atom},
			Description: gpackage.Description(),
			Author:      &feeds.Author{Name: "unknown"},
			Created:     time.Now(),
		}
		if len(gpackage.Versions) > 0 && len(gpackage.Versions[0].Commits) > 0 {
			lastCommit := gpackage.Versions[0].Commits[0]
			item.Author = &feeds.Author{Name: lastCommit.CommitterName}
			item.Created = lastCommit.CommitterDate
			item.Content = fmt.Sprintf("%s is now available in Gentoo on these architectures: %s. See <a href='https://gitweb.gentoo.org/repo/gentoo.git/commit/?id=%s'>Gitweb</a>",
				gpackage.Atom, gpackage.Versions[0].Keywords, lastCommit.Id)
		}
		f.Add(item)
	}
}
