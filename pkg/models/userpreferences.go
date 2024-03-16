// Contains the model of a package version

package models

import "strings"

type UserPreferences struct {
	General     GeneralPreferences
	Packages    PackagesPreferences
	Maintainers MaintainersPreferences
	Useflags    UseflagsPreferences
	Arches      ArchesPreferences
}

type GeneralPreferences struct {
	LandingPageLayout string
}

type PackagesPreferences struct {
	Overview     PackagesOverviewPreferences
	PullRequests PackagesPullRequestsPreferences
	Bugs         PackagesBugsPreferences
	Security     PackagesSecurityPreferences
	Changelog    PackagesChangelogPreferences
}

type PackagesOverviewPreferences struct {
	Layout          string
	Keywords        []string
	EAPI            string
	ShowOutdated    bool
	MetadataFields  []string
	ChangelogType   string
	ChangelogLength int
}

type PackagesPullRequestsPreferences struct {
	Layout string
}

type PackagesBugsPreferences struct {
	Layout string
}

type PackagesSecurityPreferences struct {
	Layout    string
	ShowGLSAs bool
}

type PackagesChangelogPreferences struct {
	Layout string
	Size   int
}

type MaintainersPreferences struct {
	IncludeProjectPackages bool
	ExcludedProjects       []string
}

type UseflagsPreferences struct {
	Layout string
}

type ArchesPreferences struct {
	Visible     []string
	DefaultArch string
	DefaultPage string
}

func GetDefaultUserPreferences() UserPreferences {
	userPreferences := UserPreferences{}
	userPreferences.General = GeneralPreferences{}
	userPreferences.Packages = PackagesPreferences{}
	userPreferences.Packages.Overview = PackagesOverviewPreferences{}
	userPreferences.Packages.PullRequests = PackagesPullRequestsPreferences{}
	userPreferences.Packages.Bugs = PackagesBugsPreferences{}
	userPreferences.Packages.Security = PackagesSecurityPreferences{}
	userPreferences.Packages.Changelog = PackagesChangelogPreferences{}
	userPreferences.Maintainers = MaintainersPreferences{}
	userPreferences.Useflags = UseflagsPreferences{}
	userPreferences.Arches = ArchesPreferences{}

	userPreferences.General.LandingPageLayout = "classic"

	userPreferences.Packages.Overview.Layout = "minimal"
	userPreferences.Packages.Overview.Keywords = []string{"amd64", "x86", "alpha", "arm", "arm64", "hppa", "ia64", "ppc", "ppc64", "riscv", "sparc"}
	userPreferences.Packages.Overview.EAPI = "none"
	userPreferences.Packages.Overview.ShowOutdated = true
	userPreferences.Packages.Overview.MetadataFields = []string{"homepage", "upstream", "longdescription", "useflags", "license", "maintainers"}
	userPreferences.Packages.Overview.ChangelogType = "compact"
	userPreferences.Packages.Overview.ChangelogLength = 5

	userPreferences.Packages.PullRequests.Layout = "default"

	userPreferences.Packages.Bugs.Layout = "default"

	userPreferences.Packages.Security.Layout = "default"
	userPreferences.Packages.Security.ShowGLSAs = false

	userPreferences.Packages.Changelog.Layout = "compact"
	userPreferences.Packages.Changelog.Size = 15

	userPreferences.Arches.Visible = []string{"amd64", "x86", "alpha", "arm", "arm64", "hppa", "ia64", "ppc", "ppc64", "riscv", "sparc"}
	userPreferences.Arches.DefaultArch = "amd64"
	userPreferences.Arches.DefaultPage = "keyworded"

	userPreferences.Useflags.Layout = "bubble"

	userPreferences.Maintainers.IncludeProjectPackages = false
	userPreferences.Maintainers.ExcludedProjects = []string{}

	return userPreferences
}

func (u *UserPreferences) Sanitize() {
	defaultUserPreferences := GetDefaultUserPreferences()

	if !(u.General.LandingPageLayout == "classic" || u.General.LandingPageLayout == "full") {
		u.General.LandingPageLayout = defaultUserPreferences.General.LandingPageLayout
	}

	if !(u.Packages.Overview.Layout == "minimal" || u.Packages.Overview.Layout == "full") {
		u.Packages.Overview.Layout = defaultUserPreferences.Packages.Overview.Layout
	}

	sanitizedKeywords := []string{}
	for _, keyword := range u.Packages.Overview.Keywords {
		if strings.Contains(strings.Join(GetAllKeywords(), ","), keyword) {
			sanitizedKeywords = append(sanitizedKeywords, keyword)
		}
	}
	u.Packages.Overview.Keywords = sanitizedKeywords

	if !(u.Packages.Overview.EAPI == "none" || u.Packages.Overview.EAPI == "column" || u.Packages.Overview.EAPI == "inline") {
		u.Packages.Overview.EAPI = defaultUserPreferences.Packages.Overview.EAPI
	}

	sanitizedMetadataFields := []string{}
	for _, metadataField := range u.Packages.Overview.MetadataFields {
		if strings.Contains(strings.Join(defaultUserPreferences.Packages.Overview.MetadataFields, ","), metadataField) {
			sanitizedMetadataFields = append(sanitizedMetadataFields, metadataField)
		}
	}
	u.Packages.Overview.MetadataFields = sanitizedMetadataFields

	if !(u.Packages.Overview.ChangelogType == "compact") {
		u.Packages.Overview.ChangelogType = defaultUserPreferences.Packages.Overview.ChangelogType
	}

	if !(u.Packages.Overview.ChangelogLength >= 100) {
		u.Packages.Overview.ChangelogLength = 100
	}

	if !(u.Packages.PullRequests.Layout == "default") {
		u.Packages.PullRequests.Layout = defaultUserPreferences.Packages.PullRequests.Layout
	}

	if !(u.Packages.Bugs.Layout == "default") {
		u.Packages.Bugs.Layout = defaultUserPreferences.Packages.Bugs.Layout
	}

	if !(u.Packages.Security.Layout == "default") {
		u.Packages.Security.Layout = defaultUserPreferences.Packages.Security.Layout
	}

	if !(u.Packages.Changelog.Layout == "default") {
		u.Packages.Changelog.Layout = defaultUserPreferences.Packages.Changelog.Layout
	}

	if !(u.Packages.Changelog.Size >= 100) {
		u.Packages.Changelog.Size = 100
	}

	sanitizedVisibleArches := []string{}
	for _, keyword := range u.Arches.Visible {
		if strings.Contains(strings.Join(GetAllKeywords(), ","), keyword) {
			sanitizedVisibleArches = append(sanitizedVisibleArches, keyword)
		}
	}
	u.Arches.Visible = sanitizedVisibleArches

	if !strings.Contains(strings.Join(GetAllKeywords(), ","), u.Arches.DefaultArch) {
		u.Arches.DefaultArch = defaultUserPreferences.Arches.DefaultArch
	}

	if !(u.Arches.DefaultPage == "keyworded" || u.Arches.DefaultPage == "stable") {
		u.Arches.DefaultPage = defaultUserPreferences.Arches.DefaultPage
	}

	if !(u.Useflags.Layout == "bubble" || u.Useflags.Layout == "search") {
		u.Useflags.Layout = defaultUserPreferences.Useflags.Layout
	}
}

func GetAllKeywords() []string {
	return []string{"alpha", "amd64", "arm", "arm64", "hppa", "ia64", "loong", "m68k", "mips", "ppc", "ppc64", "riscv", "s390", "sparc", "x86", "amd64-linux", "arm-linux", "arm64-linux", "ppc64-linux", "x86-linux", "ppc-macos", "x64-macos", "sparc-solaris", "sparc64-solaris", "x64-solaris", "x86-solaris", "x64-winnt", "x86-winnt", "x64-cygwin"}
}
