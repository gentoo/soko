// SPDX-License-Identifier: GPL-2.0-only

// Contains the model of a category

package models

type Category struct {
	Name                string `pg:",pk"`
	Description         string
	Packages            []*Package                  `pg:",fk:category"`
	PackagesInformation CategoryPackagesInformation `pg:",fk:name,rel:has-one"`
}

type CategoryPackagesInformation struct {
	Name           string `pg:",pk"`
	Outdated       int
	PullRequests   int
	Bugs           int
	SecurityBugs   int
	StableRequests int
}
