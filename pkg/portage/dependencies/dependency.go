package dependencies

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TODO
var WaitGroup sync.WaitGroup

var Dependencies []*models.ReverseDependency

var PackageCounter int
var DependencyCounter int
var ErrCounter int
var NewCounter int

var (
	mu sync.RWMutex
)

func AddDependency(dependency *models.ReverseDependency) {
	mu.Lock()
	defer mu.Unlock()
	Dependencies = append(Dependencies, dependency)
}

func AddtoErrorCounter(amount int) {
	mu.Lock()
	defer mu.Unlock()
	ErrCounter = ErrCounter + amount
}

func GetErrorCounter() int {
	mu.RLock() // readers lock
	defer mu.RUnlock()
	return ErrCounter
}

func FullPackageDependenciesUpdate() {

	database.Connect()
	defer database.DBCon.Close()

	var packages []models.Package
	database.DBCon.Model(&packages).Select()

	PackageCounter = 0
	NewCounter = 0

	cc := 0

	for _, gpackage := range packages {

		if cc%100 == 0 {
			logger.Info.Println(time.Now().Format(time.Kitchen) + ": " + strconv.Itoa(cc))
			WaitGroup.Wait()
		}

		WaitGroup.Add(1)
		go UpdatePackageDependencies(gpackage.Atom)

		cc++
	}

	logger.Info.Println("Waiting for go routines to finish")

	WaitGroup.Wait()

	logger.Info.Println()
	logger.Info.Println("Processed " + strconv.Itoa(PackageCounter) + " packages.")
	logger.Info.Println("Got " + strconv.Itoa(DependencyCounter) + " dependencies.")
	logger.Info.Println("Start inserting dependencies into the database")
	logger.Info.Println("Errors: " + strconv.Itoa(GetErrorCounter()))

	logger.Info.Println("---")

	// finally delete all outdated dependencies
	// TODO in future we want a better incremental update here
	deleteAllDependencies()

	counter := 0
	length := len(Dependencies)
	for _, dependency := range Dependencies {

		if counter%1000 == 0 {
			logger.Info.Println(time.Now().Format(time.Kitchen) + ": " + strconv.Itoa(counter) + " / " + strconv.Itoa(length))
		}

		database.DBCon.Model(dependency).WherePK().OnConflict("(id) DO UPDATE").Insert()
		counter++
	}

}

func UpdatePackageDependencies(atom string) {

	// reverse dependeny urls
	rdepend := "https://qa-reports.gentoo.org/output/genrdeps/rindex/" + atom
	parseDependencies(atom, rdepend, "rdepend")

	depend := "https://qa-reports.gentoo.org/output/genrdeps/dindex/" + atom
	parseDependencies(atom, depend, "depend")

	pdepend := "https://qa-reports.gentoo.org/output/genrdeps/pindex/" + atom
	parseDependencies(atom, pdepend, "pdepend")

	bdepend := "https://qa-reports.gentoo.org/output/genrdeps/bindex/" + atom
	parseDependencies(atom, bdepend, "bdepend")

	WaitGroup.Done()
}

func parseDependencies(atom, url, kind string) {

	client := http.Client{
		Timeout: 600 * time.Second,
	}

	resp, err := client.Get(url)

	if err != nil {
		logger.Error.Println(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	PackageCounter++

	rawResponse, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	rawDependencies := strings.Split(string(rawResponse), "\n")

	for _, rawDependency := range rawDependencies {

		dependencyParts := strings.Split(rawDependency, ":")

		if strings.TrimSpace(dependencyParts[0]) == "" {
			continue
		}

		condition := ""
		if len(dependencyParts) > 1 {
			condition = dependencyParts[1]
		}

		AddDependency(&models.ReverseDependency{
			Id:                       atom + "-" + kind + "-" + rawDependency,
			Atom:                     atom,
			Type:                     kind,
			ReverseDependencyAtom:    versionSpecifierToPackageAtom(dependencyParts[0]),
			ReverseDependencyVersion: dependencyParts[0],
			Condition:                condition,
		})

	}

}

func versionSpecifierToPackageAtom(versionSpecifier string) string {
	gpackage := strings.ReplaceAll(versionSpecifier, ">", "")
	gpackage = strings.ReplaceAll(gpackage, "<", "")
	gpackage = strings.ReplaceAll(gpackage, "=", "")
	gpackage = strings.ReplaceAll(gpackage, "~", "")

	gpackage = strings.Split(gpackage, ":")[0]

	versionnumber := regexp.MustCompile(`-[0-9]`)
	gpackage = versionnumber.Split(gpackage, 2)[0]

	return gpackage
}

// deleteAllPullrequests deletes all entries in the pullrequests and package to pull request table
func deleteAllDependencies() {
	var reverseDependencies []*models.ReverseDependency
	database.DBCon.Model(&reverseDependencies).Select()
	for _, reverseDependency := range reverseDependencies {
		database.DBCon.Model(reverseDependency).WherePK().Delete()
	}
}

func deleteOutdatedDependencies(newDependencies []*models.ReverseDependency) {
	var oldDependencies []*models.ReverseDependency
	database.DBCon.Model(&oldDependencies).Select()

	for index, oldDependency := range oldDependencies {

		if index % 10000 == 0 {
			fmt.Println(time.Now().Format(time.Kitchen) + ": " + strconv.Itoa(index) + " / " + strconv.Itoa(len(oldDependencies)))
		}

		found := false
		for _, newDependency := range newDependencies {
			if oldDependency.Id == newDependency.Id {
				found = true
			}
		}

		if !found {
			database.DBCon.Model(oldDependency).WherePK().Delete()
		}

	}

}
