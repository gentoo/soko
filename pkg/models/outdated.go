// Contains the model of a package

package models

type OutdatedPackages struct {
	Atom          string `pg:",pk"`
	GentooVersion string
	NewestVersion string
}
