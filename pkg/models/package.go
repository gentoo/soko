// Contains the model of a package

package models

type Package struct {
	Atom                    string `pg:",pk"`
	Category                string
	Name                    string
	Versions                []*Version `pg:",fk:atom"`
	Longdescription         string
	Maintainers             []*Maintainer
	Commits                 []*Commit `pg:"many2many:commit_to_packages,joinFK:commit_id"`
	PrecedingCommits  	    int
}

type Maintainer struct {
	Name                    string
	Email                   string
	Type                    string
	Restrict                string
}
