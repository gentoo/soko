// Entrypoint for the web application

package app

import (
	"soko/pkg/app/handler/about"
	"soko/pkg/app/handler/arches"
	"soko/pkg/app/handler/categories"
	"soko/pkg/app/handler/index"
	"soko/pkg/app/handler/packages"
	"log"
	"net/http"
	"soko/pkg/app/handler/useflags"
	"soko/pkg/config"
	"soko/pkg/database"
)

// Serve is used to serve the web application
func Serve() {

	database.Connect()
	defer database.DBCon.Close()

	http.HandleFunc("/categories", categories.Index)
	http.HandleFunc("/categories/", categories.Show)

	http.HandleFunc("/useflags/popular.json", useflags.Popular)
	http.HandleFunc("/useflags/suggest.json", useflags.Suggest)
	http.HandleFunc("/useflags/search", useflags.Search)
	http.HandleFunc("/useflags/", useflags.Show)
	http.HandleFunc("/useflags", useflags.Index)

	http.HandleFunc("/arches", arches.Index)
	http.HandleFunc("/arches/", arches.Show)

	http.HandleFunc("/about", about.Index)
	http.HandleFunc("/about/help", about.Help)
	http.HandleFunc("/about/queries", about.Queries)
	http.HandleFunc("/about/feedback", about.Feedback)
	http.HandleFunc("/about/feeds", about.Feeds)

	http.HandleFunc("/packages/search", packages.Search)
	http.HandleFunc("/packages/suggest.json", packages.Suggest)
	http.HandleFunc("/packages/added", packages.Added)
	http.HandleFunc("/packages/updated", packages.Updated)
	http.HandleFunc("/packages/stable", packages.Stabilized)
	http.HandleFunc("/packages/keyworded", packages.Keyworded)
	http.HandleFunc("/packages/", packages.Show)
	http.HandleFunc("/", index.Show)

	fs := http.StripPrefix("/assets/", http.FileServer(http.Dir("/go/src/soko/assets")))
	http.Handle("/assets/", fs)

	log.Fatal(http.ListenAndServe(":" + config.Port(), nil))

}
