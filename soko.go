// SPDX-License-Identifier: GPL-2.0-only
package main

import (
	"embed"
	"flag"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"

	"soko/pkg/app"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/portage"
	"soko/pkg/portage/anitya"
	"soko/pkg/portage/bugs"
	"soko/pkg/portage/dependencies"
	"soko/pkg/portage/github"
	"soko/pkg/portage/maintainers"
	"soko/pkg/portage/pkgcheck"
	"soko/pkg/portage/projects"
	"soko/pkg/portage/repology"
)

//go:embed assets
var staticAssets embed.FS

func main() {
	initLoggers()

	serve := flag.Bool("serve", false, "Start serving the application")
	update := flag.Bool("update", false, "Perform an incremental update of the package data")
	fullupdate := flag.Bool("fullupdate", false, "Perform a full update of the package data")
	updateOutdatedPackages := flag.Bool("update-outdated-packages", false, "Update the repology.org data of outdated packages")
	updatePkgcheckResults := flag.Bool("update-pkgcheck-results", false, "Update the qa-reports that is the pkgcheck results")
	updatePullrequests := flag.Bool("update-pullrequests", false, "Update the pull requests")
	flag.Bool("init-bugs", false, "Import all bugs, including the old ones. This is usually just done once.")
	updateBugs := flag.Bool("update-bugs", false, "Update the bugs belonging to the packages")
	updateDependencies := flag.Bool("update-dependencies", false, "Update the dependencies and reverse dependencies of the packages")
	updateProjects := flag.Bool("update-projects", false, "Update the project information")
	updateMaintainers := flag.Bool("update-maintainers", false, "Update the maintainer information")

	help := flag.Bool("help", false, "Print the usage of this application")

	flag.Parse()

	if *update {
		slog.Info("Updating package data")
		portage.Update()
	}
	if *fullupdate {
		slog.Info("Performing full update of the package data")
		portage.FullUpdate()
	}
	if *updateOutdatedPackages {
		updateOutdated()
	}
	if *updatePkgcheckResults {
		slog.Info("Updating the qa-reports that is the pkgcheck data")
		pkgcheck.UpdatePkgCheckResults()
	}
	if *updatePullrequests {
		slog.Info("Updating the pull requests data")
		github.FullUpdatePullRequests()
	}
	if *updateBugs {
		slog.Info("Updating the bugs data")
		bugs.UpdateBugs()
	}
	if *updateDependencies {
		slog.Info("Updating the dependencies data")
		dependencies.FullPackageDependenciesUpdate()
	}
	if *updateProjects {
		projects.UpdateProjects()
	}
	// updateMaintainers should always be executed last, as it is using
	// the updated bugs, pullrequests and and outdated packages
	if *updateMaintainers {
		slog.Info("Updating the maintainers data")
		maintainers.FullImport()
	}

	if *serve {
		app.Serve(staticAssets)
	}

	if *help {
		flag.PrintDefaults()
	}
}

func updateOutdated() {
	database.Connect()
	defer database.DBCon.Close()

	slog.Info("Updating the anitya data")
	anitya.UpdateAnitya()
	slog.Info("Updating the repology data")
	repology.UpdateOutdated()

	repology.UpdateCategoriesMetadata()
}

// initialize the loggers depending on whether
// config.debug is set to true
func initLoggers() {
	errorHandler, err := os.OpenFile(config.LogFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("failed to open error log file", config.LogFile(), "error:", err)
		errorHandler = os.Stderr
	}

	var handler slog.Handler
	if config.Debug() {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			AddSource:  true,
			TimeFormat: time.DateTime,
		})
	} else {
		handler = slogmulti.Fanout(
			tint.NewHandler(os.Stdout, &tint.Options{
				Level:      slog.LevelInfo,
				AddSource:  true,
				TimeFormat: time.DateTime,
				NoColor:    true,
			}),
			tint.NewHandler(errorHandler, &tint.Options{
				Level:      slog.LevelError,
				AddSource:  true,
				TimeFormat: time.DateTime,
				NoColor:    true,
			}),
		)
	}
	slog.SetLogLoggerLevel(slog.LevelInfo)
	slog.SetDefault(slog.New(handler))
}
