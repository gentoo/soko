// Contains miscellaneous utility functions

package utils

import (
	"soko/pkg/database"
	"soko/pkg/logger"
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
		logger.Error.Println("Error fetching application data")
		return models.Application{
			Id:         "latest",
			LastUpdate: time.Now(),
			LastCommit: "unknown",
			Version:    "unknown",
		}
	}
	return *applicationData
}
