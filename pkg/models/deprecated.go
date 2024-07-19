// SPDX-License-Identifier: GPL-2.0-only
// Contains the model of a package deprecated entry

package models

import "time"

type DeprecatedPackage struct {
	Versions    string `pg:",pk"`
	Author      string
	AuthorEmail string
	Date        time.Time
	Reason      string
}

type DeprecatedToVersion struct {
	Id                 string `pg:",pk"`
	DeprecatedVersions string
	VersionId          string
}
