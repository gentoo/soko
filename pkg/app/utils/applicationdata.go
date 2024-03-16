// Contains miscellaneous utility functions

package utils

import (
	"log/slog"
	"soko/pkg/database"
	"soko/pkg/models"
	"time"
)

// GetApplicationData is used to retrieve the
// application data from the database
func GetApplicationData() models.Application {
	// Select user by primary key.
	applicationData := &models.Application{Id: "latest"}
	err := database.DBCon.Model(applicationData).WherePK().Select()
	if err != nil {
		slog.Error("Failed fetching application data", slog.Any("err", err))
		return models.Application{
			Id:         "latest",
			LastUpdate: time.Now(),
			LastCommit: "unknown",
			Version:    "unknown",
		}
	}
	return *applicationData
}
