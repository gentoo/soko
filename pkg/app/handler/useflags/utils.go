// SPDX-License-Identifier: GPL-2.0-only
package useflags

import (
	"net/http"
	"soko/pkg/app/layout"

	"github.com/a-h/templ"
)

var tabs = []layout.SubTab{
	{Name: "Widely used", Link: templ.URL("/useflags"), Icon: "fa fa-line-chart mr-1"},
	{Name: "Search", Link: templ.URL("/useflags/search"), Icon: "fa fa-search mr-1"},
	{Name: "Global", Link: templ.URL("/useflags/global"), Icon: "fa fa-globe mr-1"},
	{Name: "Local", Link: templ.URL("/useflags/local"), Icon: "fa fa-map-marker mr-1"},
	{Name: "USE Expand", Link: templ.URL("/useflags/expand"), Icon: "fa fa-list mr-1"},
}

func RenderPage(w http.ResponseWriter, r *http.Request, title string, currentTab string, content templ.Component) {
	layout.TabbedLayout(title, layout.UseFlags, "USE flags", "fa fa-fw fa-sliders", "", tabs, currentTab, content).Render(r.Context(), w)
}
