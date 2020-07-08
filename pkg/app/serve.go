// Entrypoint for the web application

package app

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"log"
	"net/http"
	"soko/pkg/api/graphql/generated"
	"soko/pkg/api/graphql/graphiql"
	"soko/pkg/api/graphql/resolvers"
	"soko/pkg/app/handler/about"
	"soko/pkg/app/handler/arches"
	"soko/pkg/app/handler/categories"
	"soko/pkg/app/handler/index"
	"soko/pkg/app/handler/maintainer"
	"soko/pkg/app/handler/packages"
	"soko/pkg/app/handler/useflags"
	"soko/pkg/config"
	"soko/pkg/database"
)

// Serve is used to serve the web application
func Serve() {

	database.Connect()
	defer database.DBCon.Close()

	setRoute("/categories", categories.Index)
	setRoute("/categories.json", categories.JSONCategories)
	setRoute("/categories/", categories.Show)

	setRoute("/useflags/popular.json", useflags.Popular)
	setRoute("/useflags/suggest.json", useflags.Suggest)
	setRoute("/useflags/search", useflags.Search)
	setRoute("/useflags/global", useflags.Global)
	setRoute("/useflags/local", useflags.Local)
	setRoute("/useflags/", useflags.Show)
	setRoute("/useflags", useflags.Index)

	setRoute("/arches", arches.Index)
	setRoute("/arches/", arches.Show)

	setRoute("/about", about.Index)
	setRoute("/about/help", about.Help)
	setRoute("/about/feedback", about.Feedback)
	setRoute("/about/feeds", about.Feeds)

	setRoute("/maintainers", maintainer.Browse)
	setRoute("/maintainers/", maintainer.Browse)
	setRoute("/maintainer/", maintainer.Show)

	setRoute("/packages/search", packages.Search)
	setRoute("/packages/suggest.json", packages.Suggest)
	setRoute("/packages/resolve.json", packages.Resolve)
	setRoute("/packages/added", packages.Added)
	setRoute("/packages/updated", packages.Updated)
	setRoute("/packages/stable", packages.Stabilized)
	setRoute("/packages/keyworded", packages.Keyworded)
	setRoute("/packages/", packages.Show)
	setRoute("/", index.Show)

	setRoute("/packages/added.atom", packages.AddedFeed)
	setRoute("/packages/updated.atom", packages.UpdatedFeed)
	setRoute("/packages/keyworded.atom", packages.KeywordedFeed)
	setRoute("/packages/stable.atom", packages.StabilizedFeed)
	// Added for backwards compability
	redirect("/packages/stabilized.atom", "/packages/stable.atom")
	setRoute("/packages/search.atom", packages.SearchFeed)

	fs := http.StripPrefix("/assets/", http.FileServer(http.Dir("/go/src/soko/assets")))
	http.Handle("/assets/", fs)

	// api: graphql
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers:  &resolvers.Resolver{},
		Directives: generated.DirectiveRoot{},
		Complexity: generated.ComplexityRoot{},
	})
	srv := handler.NewDefaultServer(schema)
	srv.Use(extension.FixedComplexityLimit(300))
	http.Handle("/api/graphql/", cors(srv))

	// graphiql: api explorer
	setRoute("/api/explore/", graphiql.Show)

	log.Fatal(http.ListenAndServe(":"+config.Port(), nil))

}

// define a route using the default middleware and the given handler
func setRoute(path string, handler http.HandlerFunc) {
	http.HandleFunc(path, mw(handler))
}

func redirect(from, to string) {
	http.HandleFunc(from, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, 301)
	})
}

// mw is used as default middleware to set the default headers
func mw(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeaders(w)
		handler(w, r)
	}
}

// setDefaultHeaders sets the default headers that apply for all pages
func setDefaultHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", config.CacheControl())
}

func cors(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		h.ServeHTTP(w, r)
	}
}
