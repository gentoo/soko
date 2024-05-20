// Contains the model of a package version

package models

type UserPreferences struct {
	General     GeneralPreferences
	Maintainers MaintainersPreferences
}

type GeneralPreferences struct {
	LandingPageLayout string
}

type MaintainersPreferences struct {
	IncludeProjectPackages bool
	ExcludedProjects       []string
}

var ArchesToShow = [...]string{"amd64", "x86", "alpha", "arm", "arm64", "hppa", "ia64", "ppc", "ppc64", "riscv", "sparc"}
var AllArches = [...]string{"alpha", "amd64", "arm", "arm64", "hppa", "ia64", "mips", "ppc", "ppc64", "riscv", "s390", "sparc", "x86"}

func GetDefaultUserPreferences() UserPreferences {
	userPreferences := UserPreferences{}
	userPreferences.General = GeneralPreferences{}
	userPreferences.Maintainers = MaintainersPreferences{}

	userPreferences.General.LandingPageLayout = "classic"

	userPreferences.Maintainers.IncludeProjectPackages = false
	userPreferences.Maintainers.ExcludedProjects = []string{}

	return userPreferences
}

func (u *UserPreferences) Sanitize() {
	defaultUserPreferences := GetDefaultUserPreferences()

	if !(u.General.LandingPageLayout == "classic" || u.General.LandingPageLayout == "full") {
		u.General.LandingPageLayout = defaultUserPreferences.General.LandingPageLayout
	}
}
