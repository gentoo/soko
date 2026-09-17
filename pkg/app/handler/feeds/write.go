// SPDX-License-Identifier: GPL-2.0-only

package feeds

import (
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"

	"github.com/gorilla/feeds"
)

const preamble = xml.Header + `<?xml-stylesheet type="text/xsl" href="/assets/feed.xsl"?>` + "\n"

// WriteAtom writes the given feed as an Atom document. It replaces
// feeds.Feed.WriteAtom, which offers no way to add the stylesheet reference.
func WriteAtom(w http.ResponseWriter, feed *feeds.Feed) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")

	if _, err := io.WriteString(w, preamble); err != nil {
		slog.Error("Failed writing feed preamble", slog.Any("err", err))
		return
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode((&feeds.Atom{Feed: feed}).FeedXml()); err != nil {
		slog.Error("Failed writing atom feed", slog.Any("err", err))
	}
}
