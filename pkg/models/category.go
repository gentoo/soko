// Contains the model of a category

package models

type Category struct {
	Name        string `pg:",pk"`
	Description string
	Packages    []*Package `pg:",fk:category"`
}
