// Contains the model of a package

package models

type OutdatedSource string

const (
	OutdatedSourceRepology OutdatedSource = "repology"
	OutdatedSourceAnitya   OutdatedSource = "anitya"
)

type OutdatedPackages struct {
	Atom          string `pg:",pk"`
	GentooVersion string
	NewestVersion string
	Source        OutdatedSource
}
