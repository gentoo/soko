// Contains the model of a package version

package models

type Version struct {
	Id                     string `pg:",pk"`
	Category               string
	Package                string
	Atom                   string
	Version                string
	Slot                   string
	Subslot                string
	EAPI                   string
	Keywords               string
	Useflags             []string
	Restricts            []string
	Properties           []string
	Homepage             []string
	License                string
	Description            string
	Commits             []*Commit       `pg:"many2many:commit_to_versions,joinFK:commit_id"`
}
