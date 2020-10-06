// Contains the model of a package

package models

import "sort"

type Package struct {
	Atom                string `pg:",pk"`
	Category            string
	Name                string
	Versions            []*Version `pg:",fk:atom"`
	Longdescription     string
	Maintainers         []*Maintainer
	Upstream            Upstream
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
	// In case the maintainer type is "project", Project will point to the project
	Project Project `pg:",fk:email"`
	// In case the maintainer type is not "project", Projects will point to the projects the maintainer is member of
	Projects []*Project `pg:"many2many:maintainer_to_projects,joinFK:project_email"`
}

type MaintainerPackagesInformation struct {
	Outdated       int
	PullRequests   int
	Bugs           int
	SecurityBugs   int
	StableRequests int
}

type Upstream struct {
	RemoteIds []RemoteId
	BugsTo    []string
	Doc       []string
	Changelog []string
}

type RemoteId struct {
	Type string
	Id   string
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

func (p Package) AllBugs() []*Bug {
	allBugs := make(map[string]*Bug)

	for _, bug := range p.Bugs {
		allBugs[bug.Id] = bug
	}

	for _, version := range p.Versions {
		for _, bug := range version.Bugs {
			allBugs[bug.Id] = bug
		}
	}

	// convert to list
	allBugsList := []*Bug{}
	for _, bug := range allBugs {
		allBugsList = append(allBugsList, bug)
	}

	sort.Slice(allBugsList, func(i, j int) bool {
		return allBugsList[i].Id < allBugsList[j].Id
	})

	return allBugsList
}
