// SPDX-License-Identifier: GPL-2.0-only
package projects

import (
	"encoding/xml"
	"log/slog"
	"net/http"
	"soko/pkg/config"
	"soko/pkg/database"
	"soko/pkg/models"
	"time"
)

// UpdateProjects will update the database table that contains all projects
func UpdateProjects() {
	database.Connect()
	defer database.DBCon.Close()

	// get projects from api.gentoo.org
	projectList, err := parseProjectList()
	if err != nil {
		slog.Error("Error while parsing project list", slog.Any("err", err))
		return
	}

	var members []*models.MaintainerToProject
	membersMap := make(map[string]struct{})
	for _, project := range projectList {
		for _, member := range project.Members {
			id := member.Email + "|" + project.Email
			if _, ok := membersMap[id]; ok {
				continue
			}
			membersMap[id] = struct{}{}
			members = append(members, &models.MaintainerToProject{
				Id:              member.Email + "-" + project.Email,
				MaintainerEmail: member.Email,
				ProjectEmail:    project.Email,
			})
		}
	}

	// clean up the database
	database.TruncateTable((*models.Project)(nil))
	database.TruncateTable((*models.MaintainerToProject)(nil))

	// insert new project list
	_, err = database.DBCon.Model(&projectList).Insert()
	if err != nil {
		slog.Error("Error while inserting project list", slog.Any("err", err))
		return
	}
	_, err = database.DBCon.Model(&members).Insert()
	if err != nil {
		slog.Error("Error while inserting project members", slog.Any("err", err))
		return
	}

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
	if err != nil {
		return nil, err
	}

	uniqueProjects := make([]models.Project, 0, len(projectList.Projects))
	seen := make(map[string]struct{}, len(projectList.Projects))
	for _, project := range projectList.Projects {
		if _, ok := seen[project.Email]; !ok {
			seen[project.Email] = struct{}{}
			uniqueProjects = append(uniqueProjects, project)
		}
	}
	return uniqueProjects, nil
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "projects",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
