// Contains the model of a package mask entry

package models

import "time"

type Mask struct {
	Versions    string `pg:",pk"`
	Author      string
	AuthorEmail string
	Date        time.Time
	Reason      string
}

type MaskToVersion struct {
	Id           string `pg:",pk"`
	MaskVersions string
	VersionId    string
}
