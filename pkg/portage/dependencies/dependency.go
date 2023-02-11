package dependencies

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ulikunitz/xz"
)

var Dependencies []*models.ReverseDependency

var (
	mu sync.RWMutex
)

func FullPackageDependenciesUpdate() {

	database.Connect()
	defer database.DBCon.Close()

	dependencyCounter, err := UpdateDependencies()
	if err != nil {
		return
	}

	logger.Info.Println("Got", dependencyCounter, "dependencies.")

	// TODO in future we want a better incremental update here
	deleteAllDependencies()

	logger.Info.Println("Start inserting dependencies into the database")
	// because we removed all previous rows in table, we aren't concerned about
	// duplicates, so we can use bulk insert
	database.DBCon.Model(&Dependencies).Insert()

	updateStatus()
}

func UpdateDependencies() (int, error) {
	client := http.Client{
		Timeout: 600 * time.Second,
	}

	resp, err := client.Get("https://qa-reports.gentoo.org/output/genrdeps/rdeps.tar.xz")
	if err != nil {
		logger.Error.Println(err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Error.Printf("status code: %d", resp.StatusCode)
		return 0, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	xz, err := xz.NewReader(resp.Body)
	if err != nil {
		logger.Error.Println(err)
		return 0, err
	}

	var dependencyCounter int
	tr := tar.NewReader(xz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // end of tar archive
		}
		if err != nil {
			logger.Error.Println(err)
			return 0, err
		}
		switch hdr.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			nameParts := strings.SplitN(hdr.Name, "/", 2)

			rawResponse, err := ioutil.ReadAll(tr)
			if err != nil {
				logger.Error.Println(err)
				return 0, err
			}

			parseDependencies(string(rawResponse), nameParts[1], nameParts[0])
			dependencyCounter++
		}
	}
	return dependencyCounter, nil
}

func parseDependencies(rawResponse, atom, kind string) {
	rawDependencies := strings.Split(rawResponse, "\n")

	for _, rawDependency := range rawDependencies {

		dependencyParts := strings.Split(rawDependency, ":")

		if strings.TrimSpace(dependencyParts[0]) == "" {
			continue
		}

		condition := ""
		if len(dependencyParts) > 1 {
			condition = dependencyParts[1]
		}

		Dependencies = append(Dependencies, &models.ReverseDependency{
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

func deleteAllDependencies() {
	var reverseDependencies []*models.ReverseDependency
	err := database.DBCon.Model(&reverseDependencies).Column("id").Select()
	if err != nil {
		logger.Error.Println(err)
		return
	} else if len(reverseDependencies) == 0 {
		return
	}

	res, err := database.DBCon.Model(&reverseDependencies).WherePK().Delete()
	if err != nil {
		logger.Error.Println(err)
		return
	}
	logger.Info.Println("Deleted", res.RowsAffected(), "dependencies from the database.")
}

func deleteOutdatedDependencies(newDependencies []*models.ReverseDependency) {
	var oldDependencies []*models.ReverseDependency
	database.DBCon.Model(&oldDependencies).Select()

	for index, oldDependency := range oldDependencies {

		if index%10000 == 0 {
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

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "dependencies",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
