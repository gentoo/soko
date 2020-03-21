// Contains the model of a commit

package models

import "time"

type Commit struct {
	Id               string `pg:",pk"`
	PrecedingCommits int
	AuthorName       string
	AuthorEmail      string
	AuthorDate       time.Time
	CommitterName    string
	CommitterEmail   string
	CommitterDate    time.Time
	Message          string
	ChangedFiles     *ChangedFiles
	ChangedPackages  []*Package       `pg:"many2many:commit_to_packages,joinFK:package_atom"`
	ChangedVersions  []*Version       `pg:"many2many:commit_to_versions,joinFK:version_id"`
	KeywordChanges   []*KeywordChange `pg:",fk:commit_id"`
}

type ChangedFiles struct {
	Added    []*ChangedFile
	Modified []*ChangedFile
	Deleted  []*ChangedFile
}

type ChangedFile struct {
	Path       string
	ChangeType string
}

type KeywordChange struct {
	Id         string `pg:",pk"`
	CommitId   string
	Commit     *Commit `pg:",fk:commit_id"`
	VersionId  string
	Version    *Version `pg:",fk:version_id"`
	PackageId  string
	Package    *Package `pg:",fk:package_id"`
	Added      []string
	Stabilized []string
	All        []string
}

type CommitToPackage struct {
	Id          string `pg:",pk"`
	CommitId    string
	PackageAtom string
}

type CommitToVersion struct {
	Id        string `pg:",pk"`
	CommitId  string
	VersionId string
}
