package projects

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/logger"
	"soko/pkg/models"
	"time"
)

// UpdateProjects will update the database table that contains all projects
func UpdateProjects() {
	database.Connect()
	defer database.DBCon.Close()

	if config.Quiet() == "true" {
		log.SetOutput(ioutil.Discard)
	}

	// get projects from api.gentoo.org
	projectList, err := parseProjectList()
	if err != nil {
		logger.Error.Println("Error while parsing project list", err)
		return
	}

	var members []*models.MaintainerToProject
	for _, project := range projectList {
		for _, member := range project.Members {
			members = append(members, &models.MaintainerToProject{
				Id:              member.Email + "-" + project.Email,
				MaintainerEmail: member.Email,
				ProjectEmail:    project.Email,
			})
		}
	}

	// clean up the database
	database.TruncateTable[models.Project]("email")
	database.TruncateTable[models.MaintainerToProject]("id")

	// insert new project list
	database.DBCon.Model(&projectList).Insert()
	database.DBCon.Model(&members).Insert()

	updateStatus()
}

// parseProjectList gets the xml from api.gentoo.org and parses it
func parseProjectList() ([]models.Project, error) {
	resp, err := http.Get("https://api.gentoo.org/metastructure/projects.xml")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var projectList models.ProjectList
	err = xml.NewDecoder(resp.Body).Decode(&projectList)
	return projectList.Projects, err
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "projects",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
