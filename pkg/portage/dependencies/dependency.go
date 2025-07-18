// SPDX-License-Identifier: GPL-2.0-only
package dependencies

import (
	"archive/tar"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
)

var Dependencies []*models.ReverseDependency

func FullPackageDependenciesUpdate() {
	database.Connect()
	defer database.DBCon.Close()

	dependencyCounter, err := UpdateDependencies()
	if err != nil {
		return
	}

	slog.Info("collected dependencies", slog.Int("count", dependencyCounter))

	database.TruncateTable((*models.ReverseDependency)(nil))
	// because we removed all previous rows in table, we aren't concerned about
	// duplicates, so we can use bulk insert
	res, err := database.DBCon.Model(&Dependencies).Insert()
	if err != nil {
		slog.Error("Error during inserting dependencies", slog.Any("err", err))
	} else {
		slog.Info("Inserted dependencies", slog.Int("rows", res.RowsAffected()))
	}

	updateStatus()
}

func UpdateDependencies() (int, error) {
	client := http.Client{
		Timeout: 600 * time.Second,
	}

	resp, err := client.Get("https://qa-reports.gentoo.org/output/genrdeps/rdeps.tar.xz")
	if err != nil {
		slog.Error("Failed fetching dependencies", slog.Any("err", err))
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("Got bad status code", slog.Int("code", resp.StatusCode))
		return 0, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	xz, err := xz.NewReader(resp.Body)
	if err != nil {
		slog.Error("Failed decompressing dependencies", slog.Any("err", err))
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
			slog.Error("Failed reading dependencies tar", slog.Any("err", err))
			return 0, err
		}
		if hdr.Typeflag == tar.TypeReg {
			nameParts := strings.SplitN(hdr.Name, "/", 2)

			rawResponse, err := io.ReadAll(tr)
			if err != nil {
				slog.Error("Failed reading file from tar", slog.Any("err", err))
				return 0, err
			}
			parseDependencies(string(rawResponse), nameParts[1], nameParts[0])
			dependencyCounter++
		}
	}
	return dependencyCounter, nil
}

func parseDependencies(rawResponse, atom, kind string) {
	for _, rawDependency := range strings.Split(rawResponse, "\n") {
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

func updateStatus() {
	_, err := database.DBCon.Model(&models.Application{
		Id:         "dependencies",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed updating status", slog.Any("err", err))
	}
}
