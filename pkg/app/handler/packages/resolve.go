package packages

import (
	"encoding/json"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/v9/orm"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
)

// Show renders a template to show a given package
func Resolve(w http.ResponseWriter, r *http.Request) {

	atom := getParam(r, "atom")
	gpackage := new(models.Package)
	err := database.DBCon.Model(gpackage).
		Where("atom LIKE ?", "%"+atom).
		Relation("Versions").
		Relation("Versions.Masks").
		Relation("Commits", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("preceding_commits DESC").Limit(1), nil
		}).
		Limit(1).
		Select()

	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	sortVersionsDesc(gpackage.Versions)

	versions := getJSONVersions(gpackage)
	maintainers := getJSONMaintainers(gpackage)
	useflags := getJSONUseflag(gpackage)

	jsonPackage := Package{
		Atom:        gpackage.Atom,
		Description: gpackage.Versions[0].Description,
		Href:        "https://packages.gentoo.org/packages/" + gpackage.Atom,
		Versions:    versions,
		Herds:       []string{},
		Maintainers: maintainers,
		Use:         useflags,
		UpdatedAt:   gpackage.Commits[0].CommitterDate,
	}

	b, err := json.Marshal(jsonPackage)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}


func getParam(r *http.Request, q string) string {
	keys, ok := r.URL.Query()[q]

	if !ok || len(keys[0]) < 1 {
		return ""
	}

	return keys[0]
}
