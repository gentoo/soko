package bugs

import (
	"encoding/csv"
	"net/http"
	"regexp"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"strings"
)

func UpdateBugs(init bool) {
	UpdateSecurityBugs()
	UpdatePackagesBugs(init)

	UpdateClosedBugs()
}

func UpdateSecurityBugs() {
	importBugs("https://bugs.gentoo.org/buglist.cgi?component=Vulnerabilities&list_id=4688108&product=Gentoo%20Security&query_format=advanced&resolution=---&ctype=csv&human=1")
}

func UpdatePackagesBugs(init bool) {
	//
	// Keywording
	//
	importBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=UNCONFIRMED&bug_status=CONFIRMED&bug_status=IN_PROGRESS&component=Keywording&limit=0&list_id=4688124&product=Gentoo%20Linux&query_format=advanced&resolution=---&ctype=csv&human=1")

	//
	// Stabilization
	//
	importBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=UNCONFIRMED&bug_status=CONFIRMED&bug_status=IN_PROGRESS&component=Stabilization&limit=0&list_id=4688124&product=Gentoo%20Linux&query_format=advanced&resolution=---&ctype=csv&human=1")

	//
	// Current Packages
	//
	if init {
		importBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=UNCONFIRMED&bug_status=CONFIRMED&bug_status=IN_PROGRESS&chfield=%5BBug%20creation%5D&chfieldfrom=2000-01-01&chfieldto=2020-01-01&component=Current%20packages&limit=0&list_id=4688124&product=Gentoo%20Linux&query_format=advanced&resolution=---&ctype=csv&human=1")
	}
	importBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=UNCONFIRMED&bug_status=CONFIRMED&bug_status=IN_PROGRESS&chfield=%5BBug%20creation%5D&chfieldfrom=2020-01-01&chfieldto=2021-01-01&component=Current%20packages&limit=0&list_id=4688124&product=Gentoo%20Linux&query_format=advanced&resolution=---&ctype=csv&human=1")
}

func UpdateClosedBugs() {
	logger.Error.Println("UpdateClosedBugs")
	//
	// Security
	//
	deleteBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=RESOLVED&component=Vulnerabilities&list_id=4694466&product=Gentoo%20Security&query_format=advanced&resolution=FIXED&resolution=INVALID&resolution=WONTFIX&resolution=LATER&resolution=REMIND&resolution=DUPLICATE&resolution=WORKSFORME&resolution=CANTFIX&resolution=NEEDINFO&resolution=TEST-REQUEST&resolution=UPSTREAM&ctype=csv&human=1")

	//
	// Keywording
	//
	deleteBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=RESOLVED&component=Keywording&list_id=4694472&product=Gentoo%20Linux&query_format=advanced&resolution=FIXED&resolution=INVALID&resolution=WONTFIX&resolution=LATER&resolution=REMIND&resolution=DUPLICATE&resolution=WORKSFORME&resolution=CANTFIX&resolution=NEEDINFO&resolution=TEST-REQUEST&resolution=UPSTREAM&resolution=OBSOLETE&ctype=csv&human=1")

	//
	// Stabilization
	//
	deleteBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=RESOLVED&component=Stabilization&list_id=4694456&product=Gentoo%20Linux&query_format=advanced&resolution=FIXED&resolution=INVALID&resolution=WONTFIX&resolution=LATER&resolution=REMIND&resolution=DUPLICATE&resolution=WORKSFORME&resolution=CANTFIX&resolution=NEEDINFO&resolution=TEST-REQUEST&resolution=UPSTREAM&resolution=OBSOLETE&ctype=csv&human=1")

	//
	// Current Packages
	//
	deleteBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=RESOLVED&component=Current%20packages&list_id=4694476&product=Gentoo%20Linux&query_format=advanced&resolution=FIXED&resolution=INVALID&resolution=WONTFIX&resolution=LATER&resolution=REMIND&resolution=DUPLICATE&resolution=WORKSFORME&resolution=CANTFIX&resolution=NEEDINFO&resolution=TEST-REQUEST&resolution=UPSTREAM&resolution=OBSOLETE&ctype=csv&human=1")
}

func deleteBugs(source string) {
	database.Connect()
	defer database.DBCon.Close()

	data, err := readCSVFromUrl(source)
	if err != nil {
		logger.Error.Println(err)
	}

	for idx, row := range data {
		// skip header
		if idx == 0 || len(row) < 7 {
			continue
		}

		bug := models.Bug{
			Id: row[0],
		}

		//
		// Delete bug
		//
		_, err = database.DBCon.Model(&bug).WherePK().Delete()
		if err != nil {
			logger.Error.Println(err)
		}

		//
		// Delete Package To Bug
		//
		bugId := row[0]
		summary := row[6]
		summary = strings.Split(summary, " ")[0]
		affectedPackage := versionSpecifierToPackageAtom(summary)

		_, err = database.DBCon.Model(&models.PackageToBug{
			Id:          affectedPackage + "-" + bugId,
			PackageAtom: affectedPackage,
			BugId:       bugId,
		}).Where("bug_id = ?bug_id").Delete()
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func importBugs(source string) {

	database.Connect()
	defer database.DBCon.Close()

	data, err := readCSVFromUrl(source)
	if err != nil {
		logger.Error.Println(err)
	}

	for idx, row := range data {
		// skip header
		if idx == 0 || len(row) < 7 {
			continue
		}

		bug := models.Bug{
			Id:        row[0],
			Product:   row[1],
			Component: row[2],
			Assignee:  row[3],
			Status:    row[4],
			Summary:   row[6],
		}

		database.DBCon.Model(&bug).WherePK().OnConflict("(id) DO UPDATE").Insert()

		//
		// Insert Package To Bug
		//
		bugId := row[0]
		summary := row[6]
		summary = strings.Split(summary, " ")[0]
		affectedPackage := versionSpecifierToPackageAtom(summary)

		database.DBCon.Model(&models.PackageToBug{
			Id:          affectedPackage + "-" + bugId,
			PackageAtom: affectedPackage,
			BugId:       bugId,
		}).WherePK().OnConflict("(id) DO UPDATE").Insert()

	}

}

func readCSVFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ','
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// versionSpecifierToPackageAtom returns the package atom from a given version specifier
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
