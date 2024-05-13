// Entrypoint for the web application

package app

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
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

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
)

// Serve is used to serve the web application
func Serve() {
	database.Connect()
	defer database.DBCon.Close()

	setRoute("GET /categories", categories.Index)
	setRoute("GET /categories.json", categories.JSONCategories)
	setRoute("GET /categories/{category}", categories.ShowPackages)
	setRoute("GET /categories/{category}/bugs", categories.ShowBugs)
	setRoute("GET /categories/{category}/outdated", categories.ShowOutdated)
	setRoute("GET /categories/{category}/outdated.atom", categories.OutdatedFeed)
	setRoute("GET /categories/{category}/pull-requests", categories.ShowPullRequests)
	setRoute("GET /categories/{category}/security", categories.ShowSecurity)
	setRoute("GET /categories/{category}/stabilization", categories.ShowStabilizations)
	setRoute("GET /categories/{category}/stabilization.atom", categories.StabilizationFeed)
	setRoute("GET /categories/{category}/stabilization.json", categories.ShowStabilizationFile)
	setRoute("GET /categories/{category}/stabilization.list", categories.ShowStabilizationFile)
	setRoute("GET /categories/{category}/stabilization.xml", categories.ShowStabilizationFile)

	redirect("GET /useflags", "/useflags/popular")
	setRoute("GET /useflags/popular.json", useflags.Popular)
	setRoute("GET /useflags/suggest.json", useflags.Suggest)
	setRoute("GET /useflags/search", useflags.Search)
	setRoute("GET /useflags/global", useflags.Global)
	setRoute("GET /useflags/local", useflags.Local)
	setRoute("GET /useflags/expand", useflags.Expand)
	setRoute("GET /useflags/popular", useflags.PopularPage)
	setRoute("GET /useflags/{useflag}", useflags.Show)

	redirect("GET /arches", "/arches/amd64/keyworded")
	setRoute("GET /arches/{arch}/stable", arches.ShowStable)
	setRoute("GET /arches/{arch}/stable.atom", arches.ShowStableFeed)
	setRoute("GET /arches/{arch}/keyworded", arches.ShowKeyworded)
	setRoute("GET /arches/{arch}/keyworded.atom", arches.ShowKeywordedFeed)
	setRoute("GET /arches/{arch}/leaf-packages", arches.ShowLeafPackages)

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
	setRoute("GET /maintainer/{email}/changelog.atom", maintainer.ShowChangelogFeed)
	setRoute("GET /maintainer/{email}/info.json", maintainer.ShowInfoJson)
	setRoute("GET /maintainer/{email}/outdated", maintainer.ShowOutdated)
	setRoute("GET /maintainer/{email}/outdated.atom", maintainer.ShowOutdatedFeed)
	setRoute("GET /maintainer/{email}/pull-requests", maintainer.ShowPullRequests)
	setRoute("GET /maintainer/{email}/security", maintainer.ShowSecurity)
	setRoute("GET /maintainer/{email}/stabilization", maintainer.ShowStabilization)
	setRoute("GET /maintainer/{email}/stabilization.json", maintainer.ShowStabilizationFile)
	setRoute("GET /maintainer/{email}/stabilization.list", maintainer.ShowStabilizationFile)
	setRoute("GET /maintainer/{email}/stabilization.xml", maintainer.ShowStabilizationFile)
	setRoute("GET /maintainer/{email}/stabilization.atom", maintainer.ShowStabilizationFeed)

	setRoute("GET /packages/eapi6", packages.Eapi)
	setRoute("GET /packages/search", packages.Search)
	setRoute("GET /packages/suggest.json", packages.Suggest)
	setRoute("GET /packages/resolve.json", packages.Resolve)
	setRoute("GET /packages/added", packages.Added)
	setRoute("GET /packages/updated", packages.Updated)
	setRoute("GET /packages/stable", packages.Stabilized)
	setRoute("GET /packages/keyworded", packages.Keyworded)
	setRoute("GET /packages/{category}/{package}", packages.Show)
	setRoute("GET /packages/{category}/{package}/{pageName}", packages.Show)
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

	setRoute("/user/preferences/general/layout", user.General)
	setRoute("/user/preferences/general/reset", user.ResetGeneral)

	setRoute("/user/preferences/packages/edit", user.EditPackagesPreferences)
	setRoute("/user/preferences/packages/reset", user.ResetPackages)

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

	address := ":" + config.Port()
	slog.Info("Serving HTTP", "address", address)
	err := http.ListenAndServe(address, nil)
	slog.Error("exited server", "err", err)
	os.Exit(1)
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
