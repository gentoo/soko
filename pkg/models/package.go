// Contains the model of a package

package models

type Package struct {
	Atom                string `pg:",pk"`
	Category            string
	Name                string
	Versions            []*Version `pg:",fk:atom"`
	Longdescription     string
	Maintainers         []*Maintainer
	Upstream            []RemoteId
	Commits             []*Commit            `pg:"many2many:commit_to_packages,joinFK:commit_id"`
	PrecedingCommits    int                  `pg:",use_zero"`
	PkgCheckResults     []*PkgCheckResult    `pg:",fk:atom"`
	Outdated            []*OutdatedPackages  `pg:",fk:atom"`
	Bugs                []*Bug               `pg:"many2many:package_to_bugs,joinFK:bug_id"`
	PullRequests        []*GithubPullRequest `pg:"many2many:package_to_github_pull_requests,joinFK:github_pull_request_id"`
	ReverseDependencies []*ReverseDependency `pg:",fk:atom"`
}

type Maintainer struct {
	Email               string `pg:",pk"`
	Name                string
	Type                string
	Restrict            string
	PackagesInformation MaintainerPackagesInformation
}

type MaintainerPackagesInformation struct {
	Outdated     int
	PullRequests int
	Bugs         int
	SecurityBugs int
}

type RemoteId struct {
	Type          string
	Id            string
}

func (p Package) BuildRevDepMap() map[string]map[string]string {
	var data = map[string]map[string]string{}

	for _, dep := range p.ReverseDependencies {
		if data[dep.ReverseDependencyVersion] == nil {
			data[dep.ReverseDependencyVersion] = map[string]string{}
			data[dep.ReverseDependencyVersion]["Atom"] = dep.ReverseDependencyAtom
		}
		data[dep.ReverseDependencyVersion][dep.Type] = "true"
	}

	return data
}

func (p Package) Description() string {
	for _, version := range p.Versions {
		if version.Description != "" {
			return version.Description
		}
	}
	return p.Longdescription
}
