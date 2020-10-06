package user

import (
	b64 "encoding/base64"
	"encoding/json"
	"net/http"
	"soko/pkg/app/utils"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"strconv"
	"strings"
	"time"
)

// Preferences renders a template to show the user preferences page
func Preferences(w http.ResponseWriter, r *http.Request) {

	pageName := "general"

	if strings.HasSuffix(r.URL.Path, "/general") {
		pageName = "general"
	} else if strings.HasSuffix(r.URL.Path, "/packages") {
		pageName = "packages"
	} else if strings.HasSuffix(r.URL.Path, "/maintainers") {
		pageName = "maintainers"
	} else if strings.HasSuffix(r.URL.Path, "/useflags") {
		pageName = "useflags"
	} else if strings.HasSuffix(r.URL.Path, "/arches") {
		pageName = "arches"
	}

	var allProjects []*models.Project
	database.DBCon.Model(&allProjects).Select()

	renderUserTemplate(w, r, allProjects, pageName, "preferences")
}

func EditPackagesPreferences(w http.ResponseWriter, r *http.Request) {

	userPreferences := utils.GetUserPreferences(r)

	r.ParseForm()

	// Overview: Layout
	overviewLayout := r.Form.Get("overview-layout")
	if overviewLayout == "minimal" || overviewLayout == "full" {
		userPreferences.Packages.Overview.Layout = overviewLayout
	}

	// Overview: Keywords
	overviewKeywords := r.Form["overview-keywords"]
	userPreferences.Packages.Overview.Keywords = overviewKeywords

	// EAPI
	showEAPI := r.Form.Get("overview-eapi")
	if showEAPI == "none" || showEAPI == "column" || showEAPI == "inline" {
		userPreferences.Packages.Overview.EAPI = showEAPI
	}

	// Overview: Show Outdated
	userPreferences.Packages.Overview.ShowOutdated = r.Form.Get("overview-showOutdated") == "true"

	// Overview: Metadata fields
	overviewMetadataFields := r.Form["overview-metadata-fields"]
	userPreferences.Packages.Overview.MetadataFields = overviewMetadataFields

	// Overview: Changelog
	changelogSize, err := strconv.Atoi(r.Form.Get("overview-changelog-size"))
	if err == nil {
		if changelogSize < 100 {
			userPreferences.Packages.Overview.ChangelogLength = changelogSize
		} else {
			userPreferences.Packages.Overview.ChangelogLength = 100
		}
	}

	// Dependencies
	defaultDependenciesPage := r.Form.Get("dependencies-default-page")
	if defaultDependenciesPage == "dependencies" || defaultDependenciesPage == "reverse-dependencies" {
		userPreferences.Packages.Dependencies.Default = defaultDependenciesPage
	}

	// QA Report
	qaReportClasses := r.Form["qareport-classes"]
	excludedQAReportClasses := []int{}
	for i := 0; i <= 190; i++ {
		if !contains(qaReportClasses, strconv.Itoa(i)) {
			excludedQAReportClasses = append(excludedQAReportClasses, i)
		}
	}
	userPreferences.Packages.QAReport.ExcludedWarningClasses = excludedQAReportClasses

	// Tabs
	visibleTabs := r.Form["visible-tabs"]
	userPreferences.Packages.Tabs.Visible = visibleTabs

	//
	// Store cookie
	//
	encodedUserPreferences, err := json.Marshal(&userPreferences.Packages)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_packages", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/packages", http.StatusSeeOther)
}

func ResetPackages(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetDefaultUserPreferences()
	encodedUserPreferences, err := json.Marshal(&userPreferences.Packages)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_packages", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/packages", http.StatusSeeOther)
}

func General(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetUserPreferences(r)
	r.ParseForm()
	// landing page layout
	layout := r.Form.Get("landingpage-layout")
	if layout == "classic" || layout == "full" {
		userPreferences.General.LandingPageLayout = layout
	}
	// store cookie
	encodedUserPreferences, err := json.Marshal(&userPreferences.General)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_general", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/general", http.StatusSeeOther)
}

func ResetGeneral(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetDefaultUserPreferences()
	encodedUserPreferences, err := json.Marshal(&userPreferences.General)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_general", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/general", http.StatusSeeOther)
}

func Useflags(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetUserPreferences(r)
	r.ParseForm()
	// default use flag page
	layout := r.Form.Get("useflag-default-page")
	if layout == "bubble" || layout == "search" {
		userPreferences.Useflags.Layout = layout
	}
	// store cookie
	encodedUserPreferences, err := json.Marshal(&userPreferences.Useflags)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_useflags", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/useflags", http.StatusSeeOther)
}

func ResetUseflags(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetDefaultUserPreferences()
	encodedUserPreferences, err := json.Marshal(&userPreferences.Useflags)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_useflags", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/useflags", http.StatusSeeOther)
}

func Arches(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetUserPreferences(r)
	r.ParseForm()
	// visible arches
	visibleArches := r.Form["visible-arches"]
	userPreferences.Arches.Visible = visibleArches
	// default arch
	defaultArch := r.Form.Get("arches-default-arch")
	userPreferences.Arches.DefaultArch = defaultArch
	// default arches page
	defaultPage := r.Form.Get("arches-default-page")
	userPreferences.Arches.DefaultPage = defaultPage
	// store cookie
	encodedUserPreferences, err := json.Marshal(&userPreferences.Arches)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_arches", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/arches", http.StatusSeeOther)
}

func ResetArches(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetDefaultUserPreferences()
	encodedUserPreferences, err := json.Marshal(&userPreferences.Arches)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_arches", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/arches", http.StatusSeeOther)
}

func Maintainers(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetUserPreferences(r)
	r.ParseForm()
	// excluded projects
	excludedProjects := r.Form["excluded-projects"]
	userPreferences.Maintainers.ExcludedProjects = excludedProjects
	// include projects?
	includePackages := r.Form.Get("include-packages")
	userPreferences.Maintainers.IncludeProjectPackages = includePackages == "true"
	// store cookie
	encodedUserPreferences, err := json.Marshal(&userPreferences.Maintainers)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_maintainers", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/maintainers", http.StatusSeeOther)
}

func ResetMaintainers(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils.GetDefaultUserPreferences()
	encodedUserPreferences, err := json.Marshal(&userPreferences.Maintainers)
	if err == nil {
		sEnc := b64.StdEncoding.EncodeToString(encodedUserPreferences)
		addCookie(w, "userpref_maintainers", "/", sEnc, 365*24*time.Hour)
	}
	http.Redirect(w, r, "/user/preferences/maintainers", http.StatusSeeOther)
}

// addCookie will apply a new cookie to the response of a http request
// with the key/value specified.
func addCookie(w http.ResponseWriter, name, path, value string, ttl time.Duration) {
	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		Name:     name,
		Path:     path,
		Value:    value,
		Expires:  expire,
		HttpOnly: true,
		Secure:   config.DevMode() == "false",
	}
	http.SetCookie(w, &cookie)
}
