package packages

import (
	"encoding/json"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/v9/orm"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
	"time"
)

// build the json for the package
func buildJson(w http.ResponseWriter, r *http.Request) {

	atom := getAtom(r)
	gpackage := new(models.Package)
	err := database.DBCon.Model(gpackage).
		Where("atom = ?", atom).
		Relation("Versions").
		Relation("Versions.Masks").
		Relation("Commits", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("preceding_commits DESC").Limit(1), nil
		}).
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

// get all ebuild versions of the package in a format
// that is intended to be used to convert it to json
func getJSONVersions(gpackage *models.Package) []Version {
	var versions []Version
	for _, gversion := range gpackage.Versions {
		var masks []Mask
		for _, versionMask := range gversion.Masks {
			var maskedAtoms string
			if !strings.Contains(maskedAtoms, versionMask.Versions) {
				maskedAtoms = maskedAtoms + " " + versionMask.Versions
			}
			masks = append(masks, Mask{
				Author: strings.TrimSpace(versionMask.Author),
				Date:   versionMask.Date,
				Reason: strings.TrimSpace(versionMask.Reason),
				Atoms:  strings.Split(strings.TrimSpace(maskedAtoms), " "),
				Arches: []string{"*"},
			})
		}
		versions = append(versions, Version{
			Version:  gversion.Version,
			Keywords: strings.Split(gversion.Keywords, " "),
			Masks:    masks,
		})
	}
	return versions
}

// get all maintainers of the package in a format
// that is intended to be used to convert it to json
func getJSONMaintainers(gpackage *models.Package) []Maintainer {
	var maintainers []Maintainer
	for _, gmaintainers := range gpackage.Maintainers {
		maintainers = append(maintainers, Maintainer{
			Email:       gmaintainers.Email,
			Name:        gmaintainers.Name,
			Description: "",
			Type:        gmaintainers.Type,
			Members:     []Member{},
		})
	}
	return maintainers
}

// get all useflags in a format that is
// intended to be used to convert it to json
func getJSONUseflag(gpackage *models.Package) Use {
	useflags := Use{
		Local:     []Useflag{},
		Global:    []Useflag{},
		UseExpand: []Useflag{},
	}
	localUseflags, globalUseflags, useExpands := getPackageUseflags(gpackage)
	for _, useflag := range localUseflags {
		useflags.Local = append(useflags.Local, Useflag{
			Name:        useflag.Name,
			Description: useflag.Description,
		})
	}
	for _, useflag := range globalUseflags {
		useflags.Global = append(useflags.Global, Useflag{
			Name:        useflag.Name,
			Description: useflag.Description,
		})
	}
	for _, useflag := range useExpands {
		useflags.UseExpand = append(useflags.UseExpand, Useflag{
			Name:        useflag.Name,
			Description: useflag.Description,
		})
	}
	return useflags
}

type Package struct {
	Atom        string       `json:"atom"`
	Description string       `json:"description"`
	Href        string       `json:"href"`
	Versions    []Version    `json:"versions"`
	Herds       []string     `json:"herds"`
	Maintainers []Maintainer `json:"maintainers"`
	Use         Use          `json:"use"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Version struct {
	Version  string   `json:"version"`
	Keywords []string `json:"keywords"`
	Masks    []Mask   `json:"masks,omitempty"`
}

type Mask struct {
	Author string    `json:"author"`
	Date   time.Time `json:"date"`
	Reason string    `json:"reason"`
	Atoms  []string  `json:"atom"`
	Arches []string  `json:"arches"`
}

type Maintainer struct {
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Members     []Member `json:"members"`
}

type Member struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Use struct {
	Local     []Useflag `json:"local"`
	Global    []Useflag `json:"global"`
	UseExpand []Useflag `json:"use_expand"`
}

type Useflag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
