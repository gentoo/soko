package utils

import (
	b64 "encoding/base64"
	"encoding/json"
	"net/http"
	"soko/pkg/models"
)

func GetDefaultUserPreferences() models.UserPreferences {
	return models.GetDefaultUserPreferences()
}

func GetUserPreferences(r *http.Request) models.UserPreferences {
	userPreferences := models.GetDefaultUserPreferences()

	var cookie, err = r.Cookie("userpref_general")
	if err == nil {
		cookieValue, err := b64.StdEncoding.DecodeString(cookie.Value)
		if err == nil {
			json.Unmarshal(cookieValue, &userPreferences.General)
		}
	}

	cookie, err = r.Cookie("userpref_packages")
	if err == nil {
		cookieValue, err := b64.StdEncoding.DecodeString(cookie.Value)
		if err == nil {
			json.Unmarshal(cookieValue, &userPreferences.Packages)
		}
	}

	cookie, err = r.Cookie("userpref_maintainers")
	if err == nil {
		cookieValue, err := b64.StdEncoding.DecodeString(cookie.Value)
		if err == nil {
			json.Unmarshal(cookieValue, &userPreferences.Maintainers)
		}
	}

	cookie, err = r.Cookie("userpref_useflags")
	if err == nil {
		cookieValue, err := b64.StdEncoding.DecodeString(cookie.Value)
		if err == nil {
			json.Unmarshal(cookieValue, &userPreferences.Useflags)
		}
	}

	// old cookie: userpref_arches

	userPreferences.Sanitize()

	return userPreferences
}
