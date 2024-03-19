package utils

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"soko/pkg/models"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/feeds"
)

type stabilization struct {
	XMLName  xml.Name `xml:"stabilization" json:"-"`
	Category string   `xml:"category" json:"category"`
	Package  string   `xml:"package" json:"package"`
	Version  string   `xml:"version" json:"version"`
	Message  string   `xml:"message" json:"message"`
}

func (s stabilization) String() string {
	return s.Category + "/" + s.Package + "-" + s.Version + " # " + s.Message
}

func StabilizationExport(w http.ResponseWriter, pageUrl string, gpackages []*models.Package) {
	result := make([]stabilization, 0)
	for _, gpackage := range gpackages {
		for _, version := range gpackage.Versions {
			for _, pkgcheck := range version.PkgCheckResults {
				result = append(result, stabilization{
					Category: pkgcheck.Category,
					Package:  pkgcheck.Package,
					Version:  pkgcheck.Version,
					Message:  pkgcheck.Message,
				})
			}
		}
	}

	_, extension, _ := strings.Cut(pageUrl, ".")
	switch extension {
	case "json":
		b, err := json.Marshal(result)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	case "xml":
		b, err := xml.Marshal(struct {
			XMLName  xml.Name `xml:"xml"`
			Packages []stabilization
		}{Packages: result})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Write(b)
	case "list":
		var lines string
		for _, pkg := range result {
			lines += pkg.String() + "\n"
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(lines))
	}
}

func StabilizationFeed(w http.ResponseWriter, link, title string, results []*models.PkgCheckResult) {
	feed := &feeds.Feed{
		Title:   "Stabilization candidates for " + title,
		Author:  &feeds.Author{Name: "Gentoo Packages Database"},
		Created: time.Now(),
		Link:    &feeds.Link{Href: link},
	}

	for _, pkgcheck := range results {
		feed.Add(&feeds.Item{
			Title:       pkgcheck.CPV,
			Description: templ.EscapeString(pkgcheck.Message),
			Link:        &feeds.Link{Href: "https://packages.gentoo.org/packages/" + pkgcheck.Atom, Type: "text/html", Rel: "alternate"},
			Id:          pkgcheck.CPV,
		})
	}
	feed.WriteAtom(w)
}
