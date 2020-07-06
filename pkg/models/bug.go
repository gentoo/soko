package models

type Bug struct {
	Id        string `pg:",pk"`
	Product   string
	Component string
	Assignee  string
	Status    string
	Summary   string
}

type PackageToBug struct {
	Id          string `pg:",pk"`
	PackageAtom string
	BugId       string
}
