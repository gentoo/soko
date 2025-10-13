// SPDX-License-Identifier: GPL-2.0-only
package repository

import (
	"log/slog"
	"strings"

	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"soko/pkg/portage/utils"
)

func UpdatePkgMoves(paths []string) {
	pkgMoves := make(map[string]*models.PkgMove)

	for _, path := range paths {
		status, changedFile, twoParts := strings.Cut(path, "\t")
		if !twoParts {
			changedFile = path
		} else if status == "D" {
			continue
		}
		if !strings.HasPrefix(changedFile, "profiles/updates/") {
			continue
		}

		lines, err := utils.ReadLines(config.PortDir() + "/" + changedFile)
		if err != nil {
			slog.Error("Failed reading pkg moves file", slog.Any("err", err), slog.String("file", changedFile))
			continue
		}

		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) != 3 || parts[0] != "move" {
				continue
			}

			src, dst := parts[1], parts[2]
			pkgMoves[src] = &models.PkgMove{Source: src, Destination: dst}
		}
	}

	if len(pkgMoves) > 0 {
		rows := make([]*models.PkgMove, 0, len(pkgMoves))
		for _, row := range pkgMoves {
			rows = append(rows, row)
		}
		res, err := database.DBCon.Model(&rows).OnConflict("(source) DO UPDATE").Insert()
		if err != nil {
			slog.Error("Failed updating pkg moves", slog.Any("err", err))
		} else {
			slog.Info("Updated pkg moves", slog.Int("rows", res.RowsAffected()))
		}
	}
}
