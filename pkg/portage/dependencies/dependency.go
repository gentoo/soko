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

	"github.com/go-pg/pg"
	"github.com/ulikunitz/xz"
)

func FullPackageDependenciesUpdate() {
	database.Connect()
	defer database.DBCon.Close()

	update := models.Application{Id: "dependencies"}
	err := database.DBCon.Model(&update).WherePK().Select()
	if err != nil && err != pg.ErrNoRows {
		slog.Error("Failed to fetch last update time for dependencies", slog.Any("err", err))
		return
	}

	newLastModified, dependencies, err := UpdateDependencies(update.LastCommit)
	if err != nil {
		return
	} else if len(dependencies) == 0 {
		slog.Info("No new dependencies to update")
		return
	}

	slog.Info("collected dependencies", slog.Int("count", len(dependencies)))

	database.TruncateTable((*models.ReverseDependency)(nil))
	// because we removed all previous rows in table, we aren't concerned about
	// duplicates, so we can use bulk insert
	res, err := database.DBCon.Model(&dependencies).Insert()
	if err != nil {
		slog.Error("Error during inserting dependencies", slog.Any("err", err))
	} else {
		slog.Info("Inserted dependencies", slog.Int("rows", res.RowsAffected()))
	}

	updateStatus(newLastModified)
}

func UpdateDependencies(lastModified string) (newLastModified string, dependencies []*models.ReverseDependency, err error) {
	client := http.Client{
		Timeout: 600 * time.Second,
	}

	req, _ := http.NewRequest("GET", "https://qa-reports.gentoo.org/output/genrdeps/rdeps.tar.xz", nil)
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed fetching dependencies", slog.Any("err", err))
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		slog.Info("Dependencies are up to date", slog.String("lastModified", lastModified))
		return "", nil, nil
	} else if resp.StatusCode != 200 {
		slog.Error("Got bad status code", slog.Int("code", resp.StatusCode))
		return "", nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	newLastModified = resp.Header.Get("Last-Modified")

	xz, err := xz.NewReader(resp.Body)
	if err != nil {
		slog.Error("Failed decompressing dependencies", slog.Any("err", err))
		return "", nil, err
	}

	tr := tar.NewReader(xz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // end of tar archive
		}
		if err != nil {
			slog.Error("Failed reading dependencies tar", slog.Any("err", err))
			return "", nil, err
		}
		if hdr.Typeflag == tar.TypeReg {
			nameParts := strings.SplitN(hdr.Name, "/", 2)

			rawResponse, err := io.ReadAll(tr)
			if err != nil {
				slog.Error("Failed reading file from tar", slog.Any("err", err))
				return "", nil, err
			}

			atom, kind := nameParts[1], nameParts[0]

			for rawDependency := range strings.SplitSeq(string(rawResponse), "\n") {
				dependencyParts := strings.Split(rawDependency, ":")

				if strings.TrimSpace(dependencyParts[0]) == "" {
					continue
				}

				condition := ""
				if len(dependencyParts) > 1 {
					condition = dependencyParts[1]
				}

				dependencies = append(dependencies, &models.ReverseDependency{
					Id:                       atom + "-" + kind + "-" + rawDependency,
					Atom:                     atom,
					Type:                     kind,
					ReverseDependencyAtom:    versionSpecifierToPackageAtom(dependencyParts[0]),
					ReverseDependencyVersion: dependencyParts[0],
					Condition:                condition,
				})
			}
		}
	}
	return newLastModified, dependencies, nil
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

func updateStatus(lastModified string) {
	_, err := database.DBCon.Model(&models.Application{
		Id:         "dependencies",
		LastUpdate: time.Now(),
		LastCommit: lastModified,
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		slog.Error("Failed updating status", slog.Any("err", err))
	}
}
