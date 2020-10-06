// Contains the model of a package version

package models

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
	Dependencies PackagesDependenciesPreferences
	QAReport     PackagesQAReportPreferences
	PullRequests PackagesPullRequestsPreferences
	Bugs         PackagesBugsPreferences
	Security     PackagesSecurityPreferences
	Changelog    PackagesChangelogPreferences
	Tabs         PackagesTabsPreferences
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

type PackagesDependenciesPreferences struct {
	Default string
}

type PackagesQAReportPreferences struct {
	ExcludedWarningClasses []int
	ShowAll                bool
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

type PackagesTabsPreferences struct {
	Visible []string
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
	userPreferences.Packages.Dependencies = PackagesDependenciesPreferences{}
	userPreferences.Packages.QAReport = PackagesQAReportPreferences{}
	userPreferences.Packages.PullRequests = PackagesPullRequestsPreferences{}
	userPreferences.Packages.Bugs = PackagesBugsPreferences{}
	userPreferences.Packages.Security = PackagesSecurityPreferences{}
	userPreferences.Packages.Changelog = PackagesChangelogPreferences{}
	userPreferences.Packages.Tabs = PackagesTabsPreferences{}
	userPreferences.Maintainers = MaintainersPreferences{}
	userPreferences.Useflags = UseflagsPreferences{}
	userPreferences.Arches = ArchesPreferences{}

	userPreferences.General.LandingPageLayout = "classic"

	userPreferences.Packages.Overview.Layout = "minimal"
	userPreferences.Packages.Overview.Keywords = []string{"amd64", "x86", "alpha", "arm", "arm64", "hppa", "ia64", "ppc", "ppc64", "sparc"}
	userPreferences.Packages.Overview.EAPI = "none"
	userPreferences.Packages.Overview.ShowOutdated = true
	userPreferences.Packages.Overview.MetadataFields = []string{"homepage", "upstream", "longdescription", "useflags", "license", "maintainers"}
	userPreferences.Packages.Overview.ChangelogType = "compact"
	userPreferences.Packages.Overview.ChangelogLength = 5

	userPreferences.Packages.Dependencies.Default = "dependencies"

	userPreferences.Packages.QAReport.ExcludedWarningClasses = []int{}
	userPreferences.Packages.QAReport.ShowAll = true

	userPreferences.Packages.PullRequests.Layout = "default"

	userPreferences.Packages.Bugs.Layout = "default"

	userPreferences.Packages.Security.Layout = "default"
	userPreferences.Packages.Security.ShowGLSAs = false

	userPreferences.Packages.Changelog.Layout = "compact"
	userPreferences.Packages.Changelog.Size = 15

	userPreferences.Packages.Tabs.Visible = []string{"Overview", "Dependencies", "QA report", "Pull requests", "Bugs", "Security", "Changelog"}

	userPreferences.Arches.Visible = []string{"amd64", "x86", "alpha", "arm", "arm64", "hppa", "ia64", "ppc", "ppc64", "sparc"}
	userPreferences.Arches.DefaultArch = "amd64"
	userPreferences.Arches.DefaultPage = "keyworded"

	userPreferences.Useflags.Layout = "bubble"

	userPreferences.Maintainers.IncludeProjectPackages = false
	userPreferences.Maintainers.ExcludedProjects = []string{}

	return userPreferences
}

func (u UserPreferences) ContainsPkgcheckClass(class string) bool {
	for _, v := range u.Packages.QAReport.ExcludedWarningClasses {
		if GetPkgcheckClass(v) == class {
			return false
		}
	}
	return true
}

func (u UserPreferences) GetAllKeywords() []string {
	return []string{"alpha", "amd64", "arm", "arm64", "hppa", "ia64", "m68k", "mips", "ppc", "ppc64", "riscv", "s390", "sparc", "x86", "ppc-aix", "amd64-linux", "arm-linux", "arm64-linux", "ppc64-linux", "x86-linux", "ppc-macos", "x86-macos", "x64-macos", "m68k-mint", "sparc-solaris", "sparc64-solaris", "x64-solaris", "x86-solaris", "x64-winnt", "x86-winnt", "x64-cygwin", "x86-cygwin"}
}

func GetPkgcheckClass(number int) string {
	pkgcheckClasses := []string{"AbsoluteSymlink", "ArchesWithoutProfiles", "BadCommitSummary", "BadDependency", "BadDescription", "BadFilename", "BadHomepage", "BadKeywords", "BadPackageUpdate", "BadProtocol", "BadWhitespaceCharacter", "BannedCharacter", "BannedEapi", "BannedEapiCommand", "BinaryFile", "CatBadlyFormedXml", "CatInvalidXml", "CatMetadataXmlEmptyElement", "CatMetadataXmlIndentation", "CatMetadataXmlInvalidCatRef", "CatMetadataXmlInvalidPkgRef", "CatMissingMetadataXml", "ConflictingAccountIdentifiers", "ConflictingChksums", "DeadUrl", "DeprecatedChksum", "DeprecatedDep", "DeprecatedEapi", "DeprecatedEapiCommand", "DeprecatedEclass", "DeprecatedInsinto", "DirectNoMaintainer", "DirectStableKeywords", "DoubleEmptyLine", "DoublePrefixInPath", "DroppedKeywords", "DroppedStableKeywords", "DroppedUnstableKeywords", "DuplicateEclassInherits", "DuplicateFiles", "DuplicateKeywords", "EbuildIncorrectCopyright", "EbuildInvalidCopyright", "EbuildInvalidLicenseHeader", "EbuildNonGentooAuthorsCopyright", "EbuildOldGentooCopyright", "EclassBashSyntaxError", "EclassIncorrectCopyright", "EclassInvalidCopyright", "EclassInvalidLicenseHeader", "EclassNonGentooAuthorsCopyright", "EclassOldGentooCopyright", "EmptyCategoryDir", "EmptyFile", "EmptyMaintainer", "EmptyPackageDir", "EmptyProject", "EqualVersions", "ExecutableFile", "HomepageInSrcUri", "HttpsUrlAvailable", "IncorrectCopyright", "InvalidBdepend", "InvalidCommitMessage", "InvalidCommitTag", "InvalidCopyright", "InvalidDepend", "InvalidEapi", "InvalidLicense", "InvalidLicenseHeader", "InvalidPN", "InvalidPdepend", "InvalidProperties", "InvalidRdepend", "InvalidRequiredUse", "InvalidRestrict", "InvalidSlot", "InvalidSrcUri", "InvalidUTF8", "InvalidUseFlags", "LaggingProfileEapi", "LaggingStable", "MaintainerWithoutProxy", "MatchingChksums", "MatchingGlobalUse", "MismatchedPN", "MismatchedPerlVersion", "MissingAccountIdentifier", "MissingChksum", "MissingLicense", "MissingLicenseFile", "MissingLicenseRestricts", "MissingManifest", "MissingPackageRevision", "MissingPythonEclass", "MissingSignOff", "MissingSlash", "MissingSlotDep", "MissingTestRestrict", "MissingUnpackerDep", "MissingUri", "MissingUseDepDefault", "MissingVirtualKeywords", "MovedPackageUpdate", "MultiMovePackageUpdate", "NoFinalNewline", "NonGentooAuthorsCopyright", "NonexistentBlocker", "NonexistentDeps", "NonexistentProfilePath", "NonexistentProjectMaintainer", "NonsolvableDepsInDev", "NonsolvableDepsInExp", "NonsolvableDepsInStable", "ObsoleteUri", "OldGentooCopyright", "OldMultiMovePackageUpdate", "OldPackageUpdate", "OutdatedBlocker", "OutsideRangeAccountIdentifier", "OverlappingKeywords", "PkgBadlyFormedXml", "PkgInvalidXml", "PkgMetadataXmlEmptyElement", "PkgMetadataXmlIndentation", "PkgMetadataXmlInvalidCatRef", "PkgMetadataXmlInvalidPkgRef", "PkgMissingMetadataXml", "PotentialGlobalUse", "PotentialLocalUse", "PotentialStable", "ProbableGlobalUse", "ProbableUseExpand", "ProfileError", "ProfileWarning", "PythonEclassError", "PythonMissingDeps", "PythonMissingRequiredUse", "PythonRuntimeDepInAnyR1", "RdependChange", "RedirectedUrl", "RedundantDodir", "RedundantLongDescription", "RedundantUriRename", "RedundantVersion", "RequiredUseDefaults", "SSLCertificateError", "SizeViolation", "SourcingError", "StableRequest", "StaleProxyMaintProject", "StaticSrcUri", "TarballAvailable", "TrailingEmptyLine", "UncheckableDep", "UnderscoreInUseFlag", "UnknownCategories", "UnknownKeywords", "UnknownLicenses", "UnknownManifest", "UnknownMirror", "UnknownPkgDirEntry", "UnknownProfilePackageKeywords", "UnknownProfilePackageUse", "UnknownProfilePackages", "UnknownProfileUse", "UnknownProperties", "UnknownRestrict", "UnknownUseFlags", "UnnecessaryLicense", "UnnecessaryManifest", "UnnecessarySlashStrip", "UnsortedKeywords", "UnstableOnly", "UnstatedIuse", "UnusedEclasses", "UnusedGlobalUse", "UnusedInMastersEclasses", "UnusedInMastersGlobalUse", "UnusedInMastersLicenses", "UnusedInMastersMirrors", "UnusedLicenses", "UnusedLocalUse", "UnusedMirrors", "UnusedProfileDirs", "VariableInHomepage", "VisibleVcsPkg", "VulnerablePackage", "WhitespaceFound", "WrongIndentFound", "WrongMaintainerType"}
	return pkgcheckClasses[number]
}

func GetPkgcheckClassIndex(class string) int {
	pkgcheckClasses := []string{"AbsoluteSymlink", "ArchesWithoutProfiles", "BadCommitSummary", "BadDependency", "BadDescription", "BadFilename", "BadHomepage", "BadKeywords", "BadPackageUpdate", "BadProtocol", "BadWhitespaceCharacter", "BannedCharacter", "BannedEapi", "BannedEapiCommand", "BinaryFile", "CatBadlyFormedXml", "CatInvalidXml", "CatMetadataXmlEmptyElement", "CatMetadataXmlIndentation", "CatMetadataXmlInvalidCatRef", "CatMetadataXmlInvalidPkgRef", "CatMissingMetadataXml", "ConflictingAccountIdentifiers", "ConflictingChksums", "DeadUrl", "DeprecatedChksum", "DeprecatedDep", "DeprecatedEapi", "DeprecatedEapiCommand", "DeprecatedEclass", "DeprecatedInsinto", "DirectNoMaintainer", "DirectStableKeywords", "DoubleEmptyLine", "DoublePrefixInPath", "DroppedKeywords", "DroppedStableKeywords", "DroppedUnstableKeywords", "DuplicateEclassInherits", "DuplicateFiles", "DuplicateKeywords", "EbuildIncorrectCopyright", "EbuildInvalidCopyright", "EbuildInvalidLicenseHeader", "EbuildNonGentooAuthorsCopyright", "EbuildOldGentooCopyright", "EclassBashSyntaxError", "EclassIncorrectCopyright", "EclassInvalidCopyright", "EclassInvalidLicenseHeader", "EclassNonGentooAuthorsCopyright", "EclassOldGentooCopyright", "EmptyCategoryDir", "EmptyFile", "EmptyMaintainer", "EmptyPackageDir", "EmptyProject", "EqualVersions", "ExecutableFile", "HomepageInSrcUri", "HttpsUrlAvailable", "IncorrectCopyright", "InvalidBdepend", "InvalidCommitMessage", "InvalidCommitTag", "InvalidCopyright", "InvalidDepend", "InvalidEapi", "InvalidLicense", "InvalidLicenseHeader", "InvalidPN", "InvalidPdepend", "InvalidProperties", "InvalidRdepend", "InvalidRequiredUse", "InvalidRestrict", "InvalidSlot", "InvalidSrcUri", "InvalidUTF8", "InvalidUseFlags", "LaggingProfileEapi", "LaggingStable", "MaintainerWithoutProxy", "MatchingChksums", "MatchingGlobalUse", "MismatchedPN", "MismatchedPerlVersion", "MissingAccountIdentifier", "MissingChksum", "MissingLicense", "MissingLicenseFile", "MissingLicenseRestricts", "MissingManifest", "MissingPackageRevision", "MissingPythonEclass", "MissingSignOff", "MissingSlash", "MissingSlotDep", "MissingTestRestrict", "MissingUnpackerDep", "MissingUri", "MissingUseDepDefault", "MissingVirtualKeywords", "MovedPackageUpdate", "MultiMovePackageUpdate", "NoFinalNewline", "NonGentooAuthorsCopyright", "NonexistentBlocker", "NonexistentDeps", "NonexistentProfilePath", "NonexistentProjectMaintainer", "NonsolvableDepsInDev", "NonsolvableDepsInExp", "NonsolvableDepsInStable", "ObsoleteUri", "OldGentooCopyright", "OldMultiMovePackageUpdate", "OldPackageUpdate", "OutdatedBlocker", "OutsideRangeAccountIdentifier", "OverlappingKeywords", "PkgBadlyFormedXml", "PkgInvalidXml", "PkgMetadataXmlEmptyElement", "PkgMetadataXmlIndentation", "PkgMetadataXmlInvalidCatRef", "PkgMetadataXmlInvalidPkgRef", "PkgMissingMetadataXml", "PotentialGlobalUse", "PotentialLocalUse", "PotentialStable", "ProbableGlobalUse", "ProbableUseExpand", "ProfileError", "ProfileWarning", "PythonEclassError", "PythonMissingDeps", "PythonMissingRequiredUse", "PythonRuntimeDepInAnyR1", "RdependChange", "RedirectedUrl", "RedundantDodir", "RedundantLongDescription", "RedundantUriRename", "RedundantVersion", "RequiredUseDefaults", "SSLCertificateError", "SizeViolation", "SourcingError", "StableRequest", "StaleProxyMaintProject", "StaticSrcUri", "TarballAvailable", "TrailingEmptyLine", "UncheckableDep", "UnderscoreInUseFlag", "UnknownCategories", "UnknownKeywords", "UnknownLicenses", "UnknownManifest", "UnknownMirror", "UnknownPkgDirEntry", "UnknownProfilePackageKeywords", "UnknownProfilePackageUse", "UnknownProfilePackages", "UnknownProfileUse", "UnknownProperties", "UnknownRestrict", "UnknownUseFlags", "UnnecessaryLicense", "UnnecessaryManifest", "UnnecessarySlashStrip", "UnsortedKeywords", "UnstableOnly", "UnstatedIuse", "UnusedEclasses", "UnusedGlobalUse", "UnusedInMastersEclasses", "UnusedInMastersGlobalUse", "UnusedInMastersLicenses", "UnusedInMastersMirrors", "UnusedLicenses", "UnusedLocalUse", "UnusedMirrors", "UnusedProfileDirs", "VariableInHomepage", "VisibleVcsPkg", "VulnerablePackage", "WhitespaceFound", "WrongIndentFound", "WrongMaintainerType"}
	for k, v := range pkgcheckClasses {
		if v == class {
			return k
		}
	}
	return -1
}

func createSlice(n int) []int {
	slice := []int{}
	for i := 0; i <= n; i++ {
		slice = append(slice, i)
	}
	return slice
}
