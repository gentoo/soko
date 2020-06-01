// Contains the model of a pkgcheckresults

package models

type PkgCheckResult struct {
	Atom     string
	Category string
	Package  string
	Version  string
	CPV      string
	Class    string
	Message  string
}
