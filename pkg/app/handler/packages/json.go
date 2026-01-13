// SPDX-License-Identifier: GPL-2.0-only
package packages

import (
	"encoding/json"
	"net/http"
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

// build the json for the package
func buildJson(w http.ResponseWriter, r *http.Request) {
	atom := getAtom(r)
	gpackage := new(models.Package)
	err := database.DBCon.Model(gpackage).
		Where("atom = ?", atom).
		Relation("Versions").
		Relation("Versions.Masks").
		Relation("Commits", func(q *pg.Query) (*pg.Query, error) {
			return q.Order("preceding_commits DESC").Limit(1), nil
		}).
		Select()

	if err == pg.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	sortVersionsDesc(gpackage.Versions)

	if len(gpackage.Versions) == 0 || len(gpackage.Commits) == 0 {
		http.NotFound(w, r)
		return
	}

	versions := getJSONVersions(gpackage)
	maintainers := getJSONMaintainers(gpackage)
	useflags, useExpand := getJSONUseflag(gpackage)

	jsonPackage := Package{
		Atom:        gpackage.Atom,
		Description: gpackage.Versions[0].Description,
		Href:        "https://packages.gentoo.org/packages/" + gpackage.Atom,
		Versions:    versions,
		Herds:       []string{},
		Maintainers: maintainers,
		Use:         useflags,
		UseExpand:   useExpand,
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
func getJSONUseflag(gpackage *models.Package) (Use, []UseExpand) {
	useFlags := Use{
		Local:     []Useflag{},
		Global:    []Useflag{},
		UseExpand: []Useflag{},
	}
	useExpand := []UseExpand{}
	localUseflags, globalUseflags, useExpands := getPackageUseflags(gpackage)
	for _, useflag := range localUseflags {
		useFlags.Local = append(useFlags.Local, Useflag{
			Name:        useflag.Name,
			Description: useflag.Description,
		})
	}
	for _, useflag := range globalUseflags {
		useFlags.Global = append(useFlags.Global, Useflag{
			Name:        useflag.Name,
			Description: useflag.Description,
		})
	}
	for expandGroup, flags := range useExpands {
		expandedFlags := make([]Useflag, 0, len(flags))
		for _, flag := range flags {
			expandedFlags = append(expandedFlags, Useflag{
				Name:        flag.Name,
				Description: flag.Description,
			})
			prefix := ""
			if len(flag.Name) > 0 && flag.Name[0] == '+' {
				prefix, flag.Name = "+", flag.Name[1:]
			}
			useFlags.UseExpand = append(useFlags.UseExpand, Useflag{
				Name:        prefix + expandGroup + "_" + flag.Name,
				Description: flag.Description,
			})
		}
		useExpand = append(useExpand, UseExpand{Name: expandGroup, Flags: expandedFlags})
	}
	return useFlags, useExpand
}

type Package struct {
	Atom        string       `json:"atom"`
	Description string       `json:"description"`
	Href        string       `json:"href"`
	Versions    []Version    `json:"versions"`
	Herds       []string     `json:"herds"`
	Maintainers []Maintainer `json:"maintainers"`
	Use         Use          `json:"use"`
	UseExpand   []UseExpand  `json:"use_expand"`
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

type UseExpand struct {
	Name  string    `json:"name"`
	Flags []Useflag `json:"flags"`
}

type Useflag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
