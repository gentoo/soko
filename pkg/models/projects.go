package models

import "encoding/xml"

type ProjectList struct {
	XMLName  xml.Name  `xml:"projects"`
	Projects []Project `xml:"project"`
}

type Project struct {
	XMLName     xml.Name `xml:"project" pg:"-"`
	Email       string   `xml:"email" pg:",pk"`
	Name        string   `xml:"name"`
	Url         string   `xml:"url"`
	Description string   `xml:"description"`
	Members     []Member `xml:"member"`
}

type Member struct {
	XMLName xml.Name `xml:"member" json:"-" pg:"-"`
	IsLead  bool     `xml:"is-lead,attr"`
	Email   string   `xml:"email"`
	Name    string   `xml:"name"`
	Role    string   `xml:"role"`
}

type MaintainerToProject struct {
	Id              string `pg:",pk"`
	MaintainerEmail string
	ProjectEmail    string
}
