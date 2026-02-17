// SPDX-License-Identifier: GPL-2.0-only

// Contains the model of a package

package models

import (
	"sort"
	"strings"
)

type Package struct {
	Atom                string `pg:",pk"`
	Category            string
	Name                string
	Versions            []*Version `pg:",fk:atom,rel:has-many"`
	Longdescription     string
	Maintainers         []*Maintainer
	Upstream            Upstream
	AnityaInfo          *AnityaInfo
	Commits             []*Commit            `pg:"many2many:commit_to_packages,join_fk:commit_id"`
	PrecedingCommits    int                  `pg:",use_zero"`
	PkgCheckResults     []*PkgCheckResult    `pg:",fk:atom,rel:has-many"`
	Outdated            []*OutdatedPackages  `pg:",fk:atom,rel:has-many"`
	Bugs                []*Bug               `pg:"many2many:package_to_bugs,join_fk:bug_id"`
	PullRequests        []*PullRequest       `pg:"many2many:package_to_pull_requests,join_fk:pull_request_id"`
	ReverseDependencies []*ReverseDependency `pg:",fk:atom,rel:has-many"`
}

type Maintainer struct {
	Email               string `pg:",pk"`
	Name                string
	Type                string
	Restrict            string
	PackagesInformation MaintainerPackagesInformation
	// In case the maintainer type is "project", Project will point to the project
	Project Project `pg:",fk:email,rel:has-one"`
	// In case the maintainer type is not "project", Projects will point to the projects the maintainer is member of
	Projects []*Project `pg:"many2many:maintainer_to_projects,join_fk:project_email"`
}

func (m *Maintainer) PrintName() string {
	if m.Name != "" {
		return m.Name
	}
	return m.Email
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

type AnityaInfo struct {
	Project string
}

type packageDepMap struct {
	Version string
	Atom    string
	Map     map[string]struct{}
}

func (p Package) BuildRevDepMap() []packageDepMap {
	var data = map[string]packageDepMap{}

	for _, dep := range p.ReverseDependencies {
		if _, found := data[dep.ReverseDependencyVersion]; !found {
			data[dep.ReverseDependencyVersion] = packageDepMap{
				Version: dep.ReverseDependencyVersion,
				Atom:    strings.ReplaceAll(dep.ReverseDependencyAtom, "[B]", ""),
				Map:     map[string]struct{}{},
			}
		}
		data[dep.ReverseDependencyVersion].Map[dep.Type] = struct{}{}
	}

	result := make([]packageDepMap, 0, len(data))
	for _, v := range data {
		result = append(result, v)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result
}

func (p Package) Description() string {
	for _, version := range p.Versions {
		if version.Description != "" {
			return version.Description
		}
	}
	return p.Longdescription
}

func (p *Package) HasVersion(version string) bool {
	for _, v := range p.Versions {
		if v.Version == version {
			return true
		}
	}
	return false
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

func (p *Package) AllUseflags() (useflags []string, defaultOn []string) {
	useflagsSet := make(map[string]struct{})
	defaultOnSet := make(map[string]struct{})
	for _, version := range p.Versions {
		for _, useflag := range version.Useflags {
			if useflag[0] == '+' {
				defaultOnSet[useflag[1:]] = struct{}{}
				useflagsSet[useflag[1:]] = struct{}{}
			} else {
				useflagsSet[useflag] = struct{}{}
			}
		}
	}

	useflags = make([]string, 0, len(useflagsSet))
	for useflag := range useflagsSet {
		useflags = append(useflags, useflag)
	}
	defaultOn = make([]string, 0, len(defaultOnSet))
	for useflag := range defaultOnSet {
		defaultOn = append(defaultOn, useflag)
	}
	return
}

type PkgMove struct {
	Source      string `pg:",pk"`
	Destination string
}
