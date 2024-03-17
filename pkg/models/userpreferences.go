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
	Overview PackagesOverviewPreferences
}

type PackagesOverviewPreferences struct {
	Layout   string
	Keywords []string
	EAPI     string
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
	userPreferences.Maintainers = MaintainersPreferences{}
	userPreferences.Useflags = UseflagsPreferences{}
	userPreferences.Arches = ArchesPreferences{}

	userPreferences.General.LandingPageLayout = "classic"

	userPreferences.Packages.Overview.Layout = "minimal"
	userPreferences.Packages.Overview.Keywords = []string{"amd64", "x86", "alpha", "arm", "arm64", "hppa", "ia64", "ppc", "ppc64", "riscv", "sparc"}
	userPreferences.Packages.Overview.EAPI = "none"

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
