// SPDX-License-Identifier: GPL-2.0-only
// USE to show popular USE flags

package useflags

import (
	"encoding/json"
	"go/types"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"

	"github.com/go-pg/pg/v10"
)

var excludePopularUseflags = []string{"test", "doc", "debug"}

// Popular shows a json encoded list of popular USE flags
func Popular(w http.ResponseWriter, r *http.Request) {
	popular := struct {
		Name     string `json:"name"`
		Useflags []struct {
			Useflag  string       `pg:"useflag" json:"name"`
			Count    int          `pg:"count" json:"size"`
			Children types.Object `pg:"-" json:"children"`
		} `json:"children"`
	}{
		Name: "flags",
	}

	err := database.DBCon.Model((*models.Version)(nil)).
		Column("useflag").
		ColumnExpr("COUNT(useflag) AS count").
		TableExpr("jsonb_array_elements_text(useflags) AS raw_useflag").
		TableExpr("REPLACE(raw_useflag,'+','') AS useflag").
		Where(`useflag NOT LIKE '%\_%'`).
		Where("useflag NOT IN (?)", pg.In(excludePopularUseflags)).
		Group("useflag").
		Order("count DESC").
		Limit(66).
		Select(&popular.Useflags)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(popular)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
