package bugs

import (
	"encoding/csv"
	"net/http"
	"regexp"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"strings"
	"time"
)

func UpdateBugs(init bool) {
	database.Connect()
	defer database.DBCon.Close()

	updateSecurityBugs()
	updatePackagesBugs(init)

	updateClosedBugs()

	logger.Info.Println("---")

	updateStatus()
}

func updateSecurityBugs() {
	logger.Info.Println("UpdateSecurityBugs")

	importBugs("https://bugs.gentoo.org/buglist.cgi?columnlist=bug_id,product,component,assigned_to,bug_status,resolution,short_desc,changeddate,cf_stabilisation_atoms&component=Vulnerabilities&list_id=4688108&product=Gentoo%20Security&query_format=advanced&resolution=---&ctype=csv&human=1")
}

func updatePackagesBugs(init bool) {
	logger.Info.Println("UpdatePackagesBugs")
	//
	// Keywording
	//
	importBugs("https://bugs.gentoo.org/buglist.cgi?columnlist=bug_id,product,component,assigned_to,bug_status,resolution,short_desc,changeddate,cf_stabilisation_atoms&bug_status=UNCONFIRMED&bug_status=CONFIRMED&bug_status=IN_PROGRESS&component=Keywording&limit=0&list_id=4688124&product=Gentoo%20Linux&query_format=advanced&resolution=---&ctype=csv&human=1")

	//
	// Stabilization
	//
	importBugs("https://bugs.gentoo.org/buglist.cgi?columnlist=bug_id,product,component,assigned_to,bug_status,resolution,short_desc,changeddate,cf_stabilisation_atoms&bug_status=UNCONFIRMED&bug_status=CONFIRMED&bug_status=IN_PROGRESS&component=Stabilization&limit=0&list_id=4688124&product=Gentoo%20Linux&query_format=advanced&resolution=---&ctype=csv&human=1")

	//
	// Current Packages
	//
	if init {
		importBugs("https://bugs.gentoo.org/buglist.cgi?columnlist=bug_id,product,component,assigned_to,bug_status,resolution,short_desc,changeddate,cf_stabilisation_atoms&bug_status=UNCONFIRMED&bug_status=CONFIRMED&bug_status=IN_PROGRESS&chfield=%5BBug%20creation%5D&chfieldfrom=2000-01-01&chfieldto=2020-01-01&component=Current%20packages&limit=0&list_id=4688124&product=Gentoo%20Linux&query_format=advanced&resolution=---&ctype=csv&human=1")
	}
	importBugs("https://bugs.gentoo.org/buglist.cgi?columnlist=bug_id,product,component,assigned_to,bug_status,resolution,short_desc,changeddate,cf_stabilisation_atoms&bug_status=UNCONFIRMED&bug_status=CONFIRMED&bug_status=IN_PROGRESS&chfield=%5BBug%20creation%5D&chfieldfrom=2020-01-01&chfieldto=2021-01-01&component=Current%20packages&limit=0&list_id=4688124&product=Gentoo%20Linux&query_format=advanced&resolution=---&ctype=csv&human=1")
}

func updateClosedBugs() {
	logger.Info.Println("UpdateClosedBugs")
	//
	// Security
	//
	deleteBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=RESOLVED&component=Vulnerabilities&list_id=4694466&order=changeddate%20DESC%2Cbug_status%2Cpriority%2Cassigned_to%2Cbug_id&product=Gentoo%20Security&query_format=advanced&resolution=FIXED&resolution=INVALID&resolution=WONTFIX&resolution=LATER&resolution=REMIND&resolution=DUPLICATE&resolution=WORKSFORME&resolution=CANTFIX&resolution=NEEDINFO&resolution=TEST-REQUEST&resolution=UPSTREAM&ctype=csv&human=1")

	//
	// Keywording
	//
	deleteBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=RESOLVED&component=Keywording&list_id=4694472&order=changeddate%20DESC%2Cbug_status%2Cpriority%2Cassigned_to%2Cbug_id&product=Gentoo%20Linux&query_format=advanced&resolution=FIXED&resolution=INVALID&resolution=WONTFIX&resolution=LATER&resolution=REMIND&resolution=DUPLICATE&resolution=WORKSFORME&resolution=CANTFIX&resolution=NEEDINFO&resolution=TEST-REQUEST&resolution=UPSTREAM&resolution=OBSOLETE&ctype=csv&human=1")

	//
	// Stabilization
	//
	deleteBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=RESOLVED&component=Stabilization&list_id=4694456&order=changeddate%20DESC%2Cbug_status%2Cpriority%2Cassigned_to%2Cbug_id&product=Gentoo%20Linux&query_format=advanced&resolution=FIXED&resolution=INVALID&resolution=WONTFIX&resolution=LATER&resolution=REMIND&resolution=DUPLICATE&resolution=WORKSFORME&resolution=CANTFIX&resolution=NEEDINFO&resolution=TEST-REQUEST&resolution=UPSTREAM&resolution=OBSOLETE&ctype=csv&human=1")

	//
	// Current Packages
	//
	deleteBugs("https://bugs.gentoo.org/buglist.cgi?bug_status=RESOLVED&component=Current%20packages&list_id=4773158&order=changeddate%20DESC%2Cbug_status%2Cpriority%2Cassigned_to%2Cbug_id&product=Gentoo%20Linux&query_format=advanced&resolution=FIXED&resolution=INVALID&resolution=WONTFIX&resolution=LATER&resolution=REMIND&resolution=DUPLICATE&resolution=WORKSFORME&resolution=CANTFIX&resolution=NEEDINFO&resolution=TEST-REQUEST&resolution=UPSTREAM&resolution=OBSOLETE&ctype=csv&human=1")
}

func deleteBugs(source string) {
	data, err := readCSVFromUrl(source)
	if err != nil {
		logger.Error.Println(err)
		return
	}

	var bugs []*models.Bug
	var pkgsBugs []*models.PackageToBug

	for idx, row := range data {
		// skip header
		if idx == 0 || len(row) < 7 {
			continue
		}

		bugs = append(bugs, &models.Bug{
			Id: row[0],
		})
		affectedPackage := versionSpecifierToPackageAtom(strings.Split(row[6], " ")[0])
		pkgsBugs = append(pkgsBugs, &models.PackageToBug{
			Id: affectedPackage + "-" + row[0],
		})
	}

	if len(bugs) == 0 {
		return
	}

	res1, err := database.DBCon.Model(&bugs).Delete()
	if err != nil {
		logger.Error.Println("Failed to delete bugs:", err)
		return
	}

	res2, err := database.DBCon.Model(&pkgsBugs).Delete()
	if err != nil {
		logger.Error.Println("Failed to delete package bugs:", err)
		return
	}
	logger.Info.Println("Deleted", res1.RowsAffected(), "bugs and", res2.RowsAffected(), "package bugs")
}

func importBugs(source string) {
	data, err := readCSVFromUrl(source)
	if err != nil {
		logger.Error.Println(err)
		return
	}

	var bugs []*models.Bug
	var verBugs []*models.VersionToBug
	var pkgsBugs []*models.PackageToBug

	for idx, row := range data {
		// skip header
		if idx == 0 || len(row) < 7 {
			continue
		}

		bugs = append(bugs, &models.Bug{
			Id:        row[0],
			Product:   row[1],
			Component: row[2],
			Assignee:  row[3],
			Status:    row[4],
			Summary:   row[6],
		})

		//
		// Insert Package To Bug
		//
		bugId := row[0]
		summary := row[6]
		if strings.TrimSpace(row[8]) != "" {
			versions := make(map[string]struct{})
			for _, gpackage := range strings.Split(row[8], "\n") {
				affectedVersions := strings.Split(gpackage, " ")[0]
				if strings.TrimSpace(affectedVersions) != "" {
					for _, version := range calculateAffectedVersions(bugId, affectedVersions) {
						versions[version.Id] = struct{}{}
					}
				}
			}
			for version := range versions {
				verBugs = append(verBugs, &models.VersionToBug{
					Id:        version + "-" + bugId,
					VersionId: version,
					BugId:     bugId,
				})
			}
		} else {
			summary, _, _ = strings.Cut(summary, " ")
			affectedPackage := versionSpecifierToPackageAtom(summary)

			pkgsBugs = append(pkgsBugs, &models.PackageToBug{
				Id:          affectedPackage + "-" + bugId,
				PackageAtom: affectedPackage,
				BugId:       bugId,
			})
		}

	}

	res1, err := database.DBCon.Model(&bugs).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		logger.Error.Println("Failed to insert bugs:", err)
		return
	}

	res2, err := database.DBCon.Model(&verBugs).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		logger.Error.Println("Failed to insert version bugs:", err)
		return
	}

	res3, err := database.DBCon.Model(&pkgsBugs).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		logger.Error.Println("Failed to insert package bugs:", err)
		return
	}

	logger.Info.Println("Inserted", res1.RowsAffected(), "bugs,", res2.RowsAffected(), "version bugs and", res3.RowsAffected(), "package bugs")
}

func calculateAffectedVersions(bugId, versionSpecifier string) []*models.Version {
	packageAtom := versionSpecifierToPackageAtom(versionSpecifier)

	if strings.HasPrefix(versionSpecifier, "=") {
		return exactVersion(versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, "<=") {
		return comparedVersions("<=", versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, "<") {
		return comparedVersions("<", versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, ">=") {
		return comparedVersions(">=", versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, ">") {
		return comparedVersions(">", versionSpecifier, packageAtom)
	} else if strings.HasPrefix(versionSpecifier, "~") {
		return allRevisions(versionSpecifier, packageAtom)
	} else if strings.Contains(versionSpecifier, ":") {
		return versionsWithSlot(versionSpecifier, packageAtom)
	} else {
		return allVersions(versionSpecifier, packageAtom)
	}
}

// comparedVersions computes and returns all versions that are >=, >, <= or < than then given version
func comparedVersions(operator, versionSpecifier, packageAtom string) (results []*models.Version) {
	versionSpecifier = strings.ReplaceAll(versionSpecifier, operator, "")
	versionSpecifier = strings.ReplaceAll(versionSpecifier, packageAtom+"-", "")
	versionSpecifier, _, _ = strings.Cut(versionSpecifier, ":")
	givenVersion := models.Version{Version: versionSpecifier}

	var versions []*models.Version
	database.DBCon.Model(&versions).
		Where("atom = ?", packageAtom).
		Select()

	for _, v := range versions {
		if operator == ">" {
			if v.GreaterThan(givenVersion) {
				results = append(results, v)
			}
		} else if operator == ">=" {
			if v.GreaterThan(givenVersion) || v.EqualTo(givenVersion) {
				results = append(results, v)
			}
		} else if operator == "<" {
			if v.SmallerThan(givenVersion) {
				results = append(results, v)
			}
		} else if operator == "<=" {
			if v.SmallerThan(givenVersion) || v.EqualTo(givenVersion) {
				results = append(results, v)
			}
		}
	}
	return
}

var revision = regexp.MustCompile(`-r[0-9]*$`)

// allRevisions returns all revisions of the given version
func allRevisions(versionSpecifier string, packageAtom string) (versions []*models.Version) {
	versionWithoutRevision := revision.Split(versionSpecifier, 1)[0]
	versionWithoutRevision = strings.ReplaceAll(versionWithoutRevision, "~", "")
	database.DBCon.Model(&versions).
		Where("id LIKE ?", versionWithoutRevision+"%").
		Column("id").Select()

	return
}

// exactVersion returns the exact version specified in the versionSpecifier
func exactVersion(versionSpecifier string, packageAtom string) (versions []*models.Version) {
	database.DBCon.Model(&versions).
		Where("id = ?", versionSpecifier).
		Column("id").Select()

	return
}

// TODO include subslot
// versionsWithSlot returns all versions with the given slot
func versionsWithSlot(versionSpecifier string, packageAtom string) (versions []*models.Version) {
	_, slot, _ := strings.Cut(versionSpecifier, ":")

	database.DBCon.Model(&versions).
		Where("atom = ?", packageAtom).
		Where("slot = ?", slot).
		Column("id").Select()

	return
}

// allVersions returns all versions of the given package
func allVersions(versionSpecifier string, packageAtom string) (versions []*models.Version) {
	database.DBCon.Model(&versions).
		Where("atom = ?", packageAtom).
		Column("id").Select()
	return
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

var versionNumber = regexp.MustCompile(`-[0-9]`)

// versionSpecifierToPackageAtom returns the package atom from a given version specifier
func versionSpecifierToPackageAtom(versionSpecifier string) string {
	gpackage := strings.ReplaceAll(versionSpecifier, ">", "")
	gpackage = strings.ReplaceAll(gpackage, "<", "")
	gpackage = strings.ReplaceAll(gpackage, "=", "")
	gpackage = strings.ReplaceAll(gpackage, "~", "")

	gpackage, _, _ = strings.Cut(gpackage, ":")

	gpackage = versionNumber.Split(gpackage, 2)[0]

	return gpackage
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "bugs",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
