package main

import (
	"flag"
	"github.com/jasonlvhit/gocron"
	"io"
	"io/ioutil"
	"os"
	"soko/pkg/app"
	"soko/pkg/config"
	"soko/pkg/logger"
	"soko/pkg/portage"
	"soko/pkg/portage/bugs"
	"soko/pkg/portage/dependencies"
	"soko/pkg/portage/github"
	"soko/pkg/portage/maintainers"
	"soko/pkg/portage/pkgcheck"
	"soko/pkg/portage/projects"
	"soko/pkg/portage/repology"
	"soko/pkg/selfcheck"
	"time"
)

func main() {

	waitForPostgres()

	errorLogFile := logger.CreateLogFile(config.LogFile())
	defer errorLogFile.Close()
	initLoggers(os.Stdout, errorLogFile)

	serve := flag.Bool("serve", false, "Start serving the application")
	selfchecks := flag.Bool("enable-selfchecks", false, "Perform selfchecks periodicals to monitor the consistency of the data")
	update := flag.Bool("update", false, "Perform an incremental update of the package data")
	fullupdate := flag.Bool("fullupdate", false, "Perform a full update of the package data")
	updateOutdatedPackages := flag.Bool("update-outdated-packages", false, "Update the repology.org data of outdated packages")
	updatePkgcheckResults := flag.Bool("update-pkgcheck-results", false, "Update the qa-reports that is the pkgcheck results")
	updatePullrequests := flag.Bool("update-pullrequests", false, "Update the pull requests")
	initBugs := flag.Bool("init-bugs", false, "Import all bugs, including the old ones. This is usually just done once.")
	updateBugs := flag.Bool("update-bugs", false, "Update the bugs belonging to the packages")
	updateDependencies := flag.Bool("update-dependencies", false, "Update the dependencies and reverse dependencies of the packages")
	updateProjects := flag.Bool("update-projects", false, "Update the project information")
	updateMaintainers := flag.Bool("update-maintainers", false, "Update the maintainer information")

	help := flag.Bool("help", false, "Print the usage of this application")

	flag.Parse()

	if *selfchecks {
		logger.Info.Println("Enabling periodical selfcheck")
		go runSelfChecks()
		selfcheck.Serve()
	}
	if *update {
		logger.Info.Println("Updating package data")
		portage.Update()
	}
	if *fullupdate {
		logger.Info.Println("Performing full update of the package data")
		portage.FullUpdate()
	}
	if *updateOutdatedPackages {
		logger.Info.Println("Updating the repology data")
		repology.UpdateOutdated()
	}
	if *updatePkgcheckResults {
		logger.Info.Println("Updating the qa-reports that is the pkgcheck data")
		pkgcheck.UpdatePkgCheckResults()
	}
	if *updatePullrequests {
		logger.Info.Println("Updating the pull requests data")
		github.FullUpdatePullRequests()
	}
	if *initBugs {
		bugs.UpdateBugs(true)
	}
	if *updateBugs {
		logger.Info.Println("Updating the bugs data")
		bugs.UpdateBugs(false)
	}
	if *updateDependencies {
		logger.Info.Println("Updating the dependencies data")
		dependencies.FullPackageDependenciesUpdate()
	}
	if *updateProjects {
		projects.UpdateProjects()
	}
	// updateMaintainers should always be executed last, as it is using
	// the updated bugs, pullrequests and and outdated packages
	if *updateMaintainers {
		logger.Info.Println("Updating the maintainers data")
		maintainers.FullImport()
	}

	if *serve {
		app.Serve()
	}

	if *help {
		flag.PrintDefaults()
	}

}

// initialize the loggers depending on whether
// config.debug is set to true
func initLoggers(infoHandler io.Writer, errorHandler io.Writer) {
	if config.Debug() == "true" {
		logger.Init(os.Stdout, infoHandler, errorHandler)
	} else {
		logger.Init(ioutil.Discard, infoHandler, errorHandler)
	}
}

// TODO this has to be solved differently
// wait for postgres to come up
func waitForPostgres() {
	time.Sleep(5 * time.Second)
}

func runSelfChecks() {
	gocron.Every(1).Hour().From(gocron.NextTick()).Do(selfcheck.AllPackages)
	<-gocron.Start()
}
