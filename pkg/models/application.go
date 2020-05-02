// Contains the model of the application data

package models

import "time"

type Application struct {
	Id         string `pg:",pk"`
	LastUpdate time.Time
	LastCommit string
	Version    string
}

type Header struct {
	Title string
	Tab   string
}
