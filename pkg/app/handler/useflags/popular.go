// USE to show popular USE flags

package useflags

import (
	"encoding/json"
	"go/types"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
	"sort"
	"strings"
)

// Popular shows a json encoded list of popular USE flags
func Popular(w http.ResponseWriter, r *http.Request) {

	var versions []models.Version
	err := database.DBCon.Model(&versions).Column("useflags").Select()
	if err != nil {
		panic(err)
	}

	dict:= make(map[string]int)
	for _ , version :=  range versions {
		for _ , useflag :=  range version.Useflags {
			if (useflag != "test" && useflag != "doc" && useflag != "debug" && len(strings.Split(useflag, "_")) < 2) {
				dict[strings.ReplaceAll(useflag, "+", "")] = dict[strings.ReplaceAll(useflag, "+", "")] + 1
			}
		}
	}

	type kv struct {
		Key   string `json:"name"`
		Value int `json:"size"`
		Children types.Object `json:"children"`
	}

	type p struct {
		Name   string `json:"name"`
		Children []kv `json:"children"`
	}

	var ss []kv
	for k, v := range dict {
		ss = append(ss, kv{k, v, nil})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})


	popular := p{
		Name:     "flags",
		Children: ss[0:66],
	}

	b, err := json.Marshal(popular)


	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
