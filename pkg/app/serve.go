// Entrypoint for the web application

package app

import (
	"log"
	"net/http"
	_ "net/http/pprof"
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
	"soko/pkg/app/handler/user"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
)

// Serve is used to serve the web application
func Serve() {

	database.Connect()
	defer database.DBCon.Close()

	setRoute("GET /categories", categories.Index)
	setRoute("GET /categories.json", categories.JSONCategories)
	setRoute("GET /categories/", categories.Show)

	setRoute("GET /useflags/popular.json", useflags.Popular)
	setRoute("GET /useflags/suggest.json", useflags.Suggest)
	setRoute("GET /useflags/search", useflags.Search)
	setRoute("GET /useflags/global", useflags.Global)
	setRoute("GET /useflags/local", useflags.Local)
	setRoute("GET /useflags/expand", useflags.Expand)
	setRoute("GET /useflags/popular", useflags.PopularPage)
	setRoute("GET /useflags", useflags.Default)
	setRoute("GET /useflags/", useflags.Show)

	setRoute("GET /arches", arches.Index)
	setRoute("GET /arches/", arches.Show)

	setRoute("GET /about", about.Index)
	setRoute("GET /about/help", about.Help)
	setRoute("GET /about/feedback", about.Feedback)
	setRoute("GET /about/feeds", about.Feeds)
	setRoute("GET /about/status", about.Status)

	setRoute("GET /maintainers", maintainer.BrowseProjects)
	redirect("GET /maintainers/gentoo-projects", "/maintainers")
	setRoute("GET /maintainers/gentoo-developers", maintainer.BrowseDevs)
	setRoute("GET /maintainers/proxied-maintainers", maintainer.BrowseProxyDevs)
	setRoute("GET /maintainer/{email}", maintainer.ShowPackages)
	setRoute("GET /maintainer/{email}/bugs", maintainer.ShowBugs)
	setRoute("GET /maintainer/{email}/changelog", maintainer.ShowChangelog)
	setRoute("GET /maintainer/{email}/info.json", maintainer.ShowInfoJson)
	setRoute("GET /maintainer/{email}/outdated", maintainer.ShowOutdated)
	setRoute("GET /maintainer/{email}/pull-requests", maintainer.ShowPullRequests)
	setRoute("GET /maintainer/{email}/security", maintainer.ShowSecurity)
	setRoute("GET /maintainer/{email}/stabilization", maintainer.ShowStabilization)
	setRoute("GET /maintainer/{email}/stabilization.json", maintainer.ShowStabilizationFile)
	setRoute("GET /maintainer/{email}/stabilization.list", maintainer.ShowStabilizationFile)
	setRoute("GET /maintainer/{email}/stabilization.xml", maintainer.ShowStabilizationFile)

	setRoute("GET /packages/search", packages.Search)
	setRoute("GET /packages/suggest.json", packages.Suggest)
	setRoute("GET /packages/resolve.json", packages.Resolve)
	setRoute("GET /packages/added", packages.Added)
	setRoute("GET /packages/updated", packages.Updated)
	setRoute("GET /packages/stable", packages.Stabilized)
	setRoute("GET /packages/keyworded", packages.Keyworded)
	setRoute("GET /packages/", packages.Show)
	setRoute("GET /{$}", index.Show)

	setRoute("GET /packages/added.atom", packages.AddedFeed)
	setRoute("GET /packages/updated.atom", packages.UpdatedFeed)
	setRoute("GET /packages/keyworded.atom", packages.KeywordedFeed)
	setRoute("GET /packages/stable.atom", packages.StabilizedFeed)
	// Added for backwards compatibility
	redirect("GET /packages/stabilized.atom", "/packages/stable.atom")
	setRoute("GET /packages/search.atom", packages.SearchFeed)

	redirect("GET /user", "/user/preferences/general")
	redirect("GET /user/preferences", "/user/preferences/general")
	setRoute("GET /user/preferences/general", user.Preferences("General"))
	setRoute("GET /user/preferences/packages", user.Preferences("Packages"))
	setRoute("GET /user/preferences/maintainers", user.Preferences("Maintainers"))
	setRoute("GET /user/preferences/useflags", user.Preferences("USE flags"))
	setRoute("GET /user/preferences/arches", user.Preferences("Architectures"))

	setRoute("/user/preferences/general/layout", user.General)
	setRoute("/user/preferences/general/reset", user.ResetGeneral)

	setRoute("/user/preferences/arches/visible", user.Arches)
	setRoute("/user/preferences/arches/reset", user.ResetArches)

	setRoute("/user/preferences/packages/edit", user.EditPackagesPreferences)
	setRoute("/user/preferences/packages/reset", user.ResetPackages)

	setRoute("/user/preferences/useflags/edit", user.Useflags)
	setRoute("/user/preferences/useflags/reset", user.ResetUseflags)

	setRoute("/user/preferences/maintainers/edit", user.Maintainers)
	setRoute("/user/preferences/maintainers/reset", user.ResetMaintainers)

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

	logger.Info.Println("Serving on port: " + config.Port())
	log.Fatal(http.ListenAndServe(":"+config.Port(), nil))
}

// define a route using the default middleware and the given handler
func setRoute(path string, handler http.HandlerFunc) {
	http.HandleFunc(path, mw(handler))
}

func redirect(from, to string) {
	http.HandleFunc(from, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, http.StatusMovedPermanently)
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
