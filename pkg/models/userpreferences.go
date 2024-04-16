// Contains the model of a package version

package models

type UserPreferences struct {
	General     GeneralPreferences
	Packages    PackagesPreferences
	Maintainers MaintainersPreferences
	Useflags    UseflagsPreferences
}

type GeneralPreferences struct {
	LandingPageLayout string
}

type PackagesPreferences struct {
	Overview PackagesOverviewPreferences
}

type PackagesOverviewPreferences struct {
	Layout string
	EAPI   string
}

type MaintainersPreferences struct {
	IncludeProjectPackages bool
	ExcludedProjects       []string
}

type UseflagsPreferences struct {
	Layout string
}

var ArchesToShow = [...]string{"amd64", "x86", "alpha", "arm", "arm64", "hppa", "ia64", "ppc", "ppc64", "riscv", "sparc"}

func GetDefaultUserPreferences() UserPreferences {
	userPreferences := UserPreferences{}
	userPreferences.General = GeneralPreferences{}
	userPreferences.Packages = PackagesPreferences{}
	userPreferences.Packages.Overview = PackagesOverviewPreferences{}
	userPreferences.Maintainers = MaintainersPreferences{}
	userPreferences.Useflags = UseflagsPreferences{}

	userPreferences.General.LandingPageLayout = "classic"

	userPreferences.Packages.Overview.Layout = "minimal"
	userPreferences.Packages.Overview.EAPI = "none"

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

	if !(u.Packages.Overview.EAPI == "none" || u.Packages.Overview.EAPI == "column" || u.Packages.Overview.EAPI == "inline") {
		u.Packages.Overview.EAPI = defaultUserPreferences.Packages.Overview.EAPI
	}

	if !(u.Useflags.Layout == "bubble" || u.Useflags.Layout == "search") {
		u.Useflags.Layout = defaultUserPreferences.Useflags.Layout
	}
}
