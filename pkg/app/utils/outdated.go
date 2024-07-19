// SPDX-License-Identifier: GPL-2.0-only
package utils

import (
	"net/http"
	"time"

	"github.com/gorilla/feeds"

	"soko/pkg/models"
)

func OutdatedFeed(w http.ResponseWriter, link, title string, outdated []models.OutdatedPackages) {
	feed := &feeds.Feed{
		Title:   "Outdated Packages for " + title,
		Author:  &feeds.Author{Name: "Gentoo Packages Database"},
		Created: time.Now(),
		Link:    &feeds.Link{Href: link},
	}

	for _, entry := range outdated {
		feed.Add(&feeds.Item{
			Id:          entry.Atom,
			Title:       entry.Atom,
			Description: "Version " + entry.NewestVersion + " is available, while the latest version in the Gentoo tree is " + entry.GentooVersion + ".",
			Link:        &feeds.Link{Href: "https://packages.gentoo.org/packages/" + entry.Atom, Type: "text/html", Rel: "alternate"},
		})
	}
	feed.WriteAtom(w)
}
