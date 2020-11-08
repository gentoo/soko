// Contains utility functions to execute commands and parse the output

package storage

import "soko/pkg/models"

var (
	Packages []*models.Package
	Versions []*models.Version
	Useflags []*models.Useflag
	Masks []*models.Mask
	Categories []*models.Category
)