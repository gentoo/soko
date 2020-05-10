// Contains the model of a package version

package models

import (
	"github.com/mcuadros/go-version"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type Version struct {
	Id          string `pg:",pk"`
	Category    string
	Package     string
	Atom        string
	Version     string
	Slot        string
	Subslot     string
	EAPI        string
	Keywords    string
	Useflags    []string
	Restricts   []string
	Properties  []string
	Homepage    []string
	License     string
	Description string
	Commits     []*Commit `pg:"many2many:commit_to_versions,joinFK:commit_id"`
	Masks       []*Mask   `pg:"many2many:mask_to_versions,joinFK:mask_versions"`
}


// Compare two versions strings - compliant to the 'Version Comparison'
// described in the Package Manager Specification (PMS)
func (v *Version) CompareTo(other Version) bool {
	versionIdentifierA := v.computeVersionIdentifier()
	versionIdentifierB := other.computeVersionIdentifier()

	// compare the numeric part
	numericPartA := version.Normalize(versionIdentifierA.NumericPart)
	numericPartB := version.Normalize(versionIdentifierB.NumericPart)
	if !version.Compare(numericPartA, numericPartB, "=") {
		return version.Compare(numericPartA, numericPartB, ">")
	}

	// compare the letter
	if versionIdentifierA.Letter != versionIdentifierB.Letter {
		return strings.Compare(versionIdentifierA.Letter, versionIdentifierB.Letter) == 1
	}

	// compare the suffixes
	for i := 0; i < min(len(versionIdentifierA.Suffixes), len(versionIdentifierB.Suffixes)); i++ {
		if versionIdentifierA.Suffixes[i] == versionIdentifierA.Suffixes[i] {
			return versionIdentifierA.Suffixes[i].Number > versionIdentifierB.Suffixes[i].Number
		} else {
			return getSuffixOrder(versionIdentifierA.Suffixes[i].Name) > getSuffixOrder(versionIdentifierB.Suffixes[i].Name)
		}
	}
	if len(versionIdentifierA.Suffixes) != len(versionIdentifierB.Suffixes) {
		return len(versionIdentifierA.Suffixes) > len(versionIdentifierB.Suffixes)
	}

	// compare the revision
	if versionIdentifierA.Revision != versionIdentifierB.Revision {
		return versionIdentifierA.Revision > versionIdentifierB.Revision
	}

	// the versions are equal based on the PMS specification but
	// we have to return a bool, that's why we return true here
	return true
}



// utils

type VersionIdentifier struct {
	NumericPart        string
	Letter             string
	Suffixes           []*VersionSuffix
	Revision           int
}

type VersionSuffix struct {
	Name        string
	Number      int
}

// get the minimum of the two given ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// computeVersionIdentifier is parsing the Version string of a
// version and is computing a VersionIdentifier based on this
// string.
func (v *Version) computeVersionIdentifier() VersionIdentifier {

	rawVersionParts := strings.FieldsFunc(v.Version, func(r rune) bool {
		return r == '_' || r == '-'
	})

	versionIdentifier := new(VersionIdentifier)
	versionIdentifier.NumericPart, versionIdentifier.Letter = getNumericPart(rawVersionParts[0])
	rawVersionParts = rawVersionParts[1:]

	for _, rawVersionPart := range rawVersionParts {
		if suffix := getSuffix(rawVersionPart); suffix != nil {
			versionIdentifier.Suffixes = append(versionIdentifier.Suffixes, suffix)
		} else if isRevision(rawVersionPart) {
			parsedRevision, err := strconv.Atoi(strings.ReplaceAll(rawVersionPart, "r", ""))
			if err == nil {
				versionIdentifier.Revision = parsedRevision
			}
		}
	}

	return *versionIdentifier
}

// getNumericPart returns the numeric part of the version, that is:
//   version, letter
// i.e. 10.3.18a becomes
//   10.3.18, a
// The first returned string is the version and the second if the (optional) letter
func getNumericPart(str string) (string, string) {
	if unicode.IsLetter(rune(str[len(str)-1])) {
		return str[:len(str)-1], str[len(str)-1:]
	}
	return str, ""
}

// getSuffix creates a VersionSuffix based on the given string.
// The given string is expected to be look like
//   pre20190518
// for instance. The suffix named as well as the following number
// will be parsed and returned as VersionSuffix
func getSuffix(str string) *VersionSuffix {
	allowedSuffixes := []string{"alpha", " beta", "pre", "rc", "p"}
	for _, allowedSuffix := range allowedSuffixes {
		if regexp.MustCompile(allowedSuffix + `\d+`).MatchString(str) {
			parsedSuffix, err := strconv.Atoi(strings.ReplaceAll(str, allowedSuffix, ""))
			if err == nil {
				return &VersionSuffix{
					Name:  allowedSuffix,
					Number: parsedSuffix,
				}
			}
		}
	}
	return nil
}

// isRevision checks whether the given string
// matches the format of a revision, that is
// 'r2' for instance.
func isRevision(str string) bool {
	return regexp.MustCompile(`r\d+`).MatchString(str)
}

// getSuffixOrder returns an int for the given suffix,
// based on the following:
//   _alpha < _beta < _pre < _rc < _p < none
// as defined in the Package Manager Specification (PMS)
func getSuffixOrder(suffix string) int {
	if suffix == "p" {
		return 4
	} else if suffix == "rc" {
		return 3
	} else if suffix == "pre" {
		return 2
	} else if suffix == "beta" {
		return 1
	} else if suffix == "alpha" {
		return 0
	} else {
		return 9999
	}
}
