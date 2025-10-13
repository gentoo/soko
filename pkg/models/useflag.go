// SPDX-License-Identifier: GPL-2.0-only

// Contains the model of a USE flag

package models

type Useflag struct {
	Id          string `pg:",pk"`
	Name        string
	Scope       string
	Description string
	UseExpand   string
	Package     string
}
