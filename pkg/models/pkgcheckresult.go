// Contains the model of a pkgcheckresults

package models

type PkgCheckResult struct {
	Id       string `pg:",pk"`
	Atom     string
	Category string
	Package  string
	Version  string
	CPV      string
	Class    string
	Message  string
}
