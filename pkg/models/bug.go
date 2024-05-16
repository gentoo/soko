package models

type BugComponent string

const (
	BugComponentVulnerabilities BugComponent = "Vulnerabilities"
	BugComponentStabilization   BugComponent = "Stabilization"
	BugComponentKeywording      BugComponent = "Keywording"
	BugComponentGeneral         BugComponent = ""
)

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

type VersionToBug struct {
	Id        string `pg:",pk"`
	VersionId string
	BugId     string
}

func (b *Bug) MatchesComponent(component BugComponent) bool {
	if component != BugComponentGeneral {
		return b.Component == string(component)
	}
	return b.Component != string(BugComponentVulnerabilities) &&
		b.Component != string(BugComponentStabilization) &&
		b.Component != string(BugComponentKeywording)
}
