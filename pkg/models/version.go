// Contains the model of a package version

package models

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type Version struct {
	Id              string `pg:",pk"`
	Category        string
	Package         string
	Atom            string
	Version         string
	Slot            string
	Subslot         string
	EAPI            string
	Keywords        string
	Useflags        []string
	Restricts       []string
	Properties      []string
	Homepage        []string
	License         string
	Description     string
	Commits         []*Commit            `pg:"many2many:commit_to_versions,joinFK:commit_id"`
	Masks           []*Mask              `pg:"many2many:mask_to_versions,joinFK:mask_versions"`
	PkgCheckResults []*PkgCheckResult    `pg:",fk:cpv"`
	Dependencies    []*ReverseDependency `pg:",fk:reverse_dependency_version"`
	Bugs            []*Bug               `pg:"many2many:version_to_bugs,joinFK:bug_id"`
}

func (v Version) BuildDepMap() map[string]map[string]string {
	var data = map[string]map[string]string{}

	for _, dep := range v.Dependencies {
		if data[dep.Atom] == nil {
			data[dep.Atom] = map[string]string{}
			data[dep.Atom]["Atom"] = dep.Atom
		}
		data[dep.Atom][dep.Type] = "true"
	}

	return data
}

// GreaterThan returns true if the version is greater than the given version
// compliant to the 'Version Comparison' described in the Package Manager Specification (PMS)
func (v *Version) GreaterThan(other Version) bool {
	versionIdentifierA := v.computeVersionIdentifier()
	versionIdentifierB := other.computeVersionIdentifier()

	if v.Version == "0.7.8_p0-r100" {
		fmt.Println("versionIdentifierA")
		fmt.Println(versionIdentifierA)
		fmt.Println("versionIdentifierB")
		fmt.Println(versionIdentifierB)
	}

	// compare the numeric part
	numericPartsA := strings.Split(versionIdentifierA.NumericPart, ".")
	numericPartsB := strings.Split(versionIdentifierB.NumericPart, ".")
	if numberGreaterThan(numericPartsA[0], numericPartsB[0]) {
		return true
	} else if numberGreaterThan(numericPartsB[0], numericPartsA[0]) {
		return false
	}
	for i := 1; i < min(len(numericPartsA), len(numericPartsB)); i++ {
		// Version comparison logic for each numeric component after the first
		if strings.HasPrefix(numericPartsA[i], "0") || strings.HasPrefix(numericPartsB[i], "0") {
			NumericPartA := strings.TrimRight(numericPartsA[i], "0")
			NumericPartB := strings.TrimRight(numericPartsB[i], "0")
			if NumericPartA > NumericPartB {
				return true
			} else if NumericPartA < NumericPartB {
				return false
			}
		} else {

			if numberGreaterThan(numericPartsA[i], numericPartsB[i]) {
				return true
			} else if numberGreaterThan(numericPartsB[i], numericPartsA[i]) {
				return false
			}
		}
	}
	if len(numericPartsA) > len(numericPartsB) {
		return true
	} else if len(numericPartsA) < len(numericPartsB) {
		return false
	}

	// compare the letter
	if versionIdentifierA.Letter != versionIdentifierB.Letter {
		return strings.Compare(versionIdentifierA.Letter, versionIdentifierB.Letter) == 1
	}

	// compare the suffixes
	for i := 0; i < min(len(versionIdentifierA.Suffixes), len(versionIdentifierB.Suffixes)); i++ {
		if versionIdentifierA.Suffixes[i].Name == versionIdentifierB.Suffixes[i].Name {
			if versionIdentifierA.Suffixes[i].Number != versionIdentifierB.Suffixes[i].Number {
				return versionIdentifierA.Suffixes[i].Number > versionIdentifierB.Suffixes[i].Number
			}
		} else {
			return getSuffixOrder(versionIdentifierA.Suffixes[i].Name) > getSuffixOrder(versionIdentifierB.Suffixes[i].Name)
		}
	}
	if len(versionIdentifierA.Suffixes) > len(versionIdentifierB.Suffixes) {
		if versionIdentifierA.Suffixes[len(versionIdentifierB.Suffixes)].Name == "p" {
			return true
		} else {
			return false
		}
	} else if len(versionIdentifierA.Suffixes) < len(versionIdentifierB.Suffixes) {
		if versionIdentifierB.Suffixes[len(versionIdentifierA.Suffixes)].Name == "p" {
			return false
		} else {
			return true
		}
	}

	// compare the revision
	if versionIdentifierA.Revision != versionIdentifierB.Revision {
		return versionIdentifierA.Revision > versionIdentifierB.Revision
	}

	// the versions are equal based on the PMS specification
	return false
}

// SmallerThan returns true if the version is smaller than the given version
// compliant to the 'Version Comparison' described in the Package Manager Specification (PMS)
func (v *Version) SmallerThan(other Version) bool {
	return other.GreaterThan(*v)
}

// EqualTo returns true if the version is equal to the given version
// compliant to the 'Version Comparison' described in the Package Manager Specification (PMS)
func (v *Version) EqualTo(other Version) bool {
	return !v.GreaterThan(other) && !v.SmallerThan(other)
}

// utils

type VersionIdentifier struct {
	NumericPart string
	Letter      string
	Suffixes    []*VersionSuffix
	Revision    int
}

type VersionSuffix struct {
	Name   string
	Number int
}

// get the minimum of the two given ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// numberGreaterThan takes two strings and returns true if the
// first strings is greater than the second one using integer
// comparison. - In case of an error during the string to int
// conversion false will be returned
func numberGreaterThan(a, b string) bool {

	aInt := new(big.Int)
	aInt, aOK := aInt.SetString(a, 10)
	bInt := new(big.Int)
	bInt, bOK := bInt.SetString(b, 10)

	if !aOK || !bOK {
		return false
	}

	return aInt.Cmp(bInt) == 1
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
	allowedSuffixes := []string{"alpha", "beta", "pre", "rc", "p"}
	for _, allowedSuffix := range allowedSuffixes {
		if regexp.MustCompile(allowedSuffix + `\d+`).MatchString(str) {
			parsedSuffix, err := strconv.Atoi(strings.ReplaceAll(str, allowedSuffix, ""))
			if err == nil {
				return &VersionSuffix{
					Name:   allowedSuffix,
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
