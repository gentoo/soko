// Contains utility functions around the database

package database

import (
	"context"
	"log"
	"soko/pkg/config"
	"soko/pkg/logger"
	"soko/pkg/models"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

// DBCon is the connection handle
// for the database
var (
	DBCon *pg.DB
)

// CreateSchema creates the tables in the database
// in case they don't already exist
func CreateSchema() error {
	for _, model := range []interface{}{
		(*models.CommitToPackage)(nil),
		(*models.CommitToVersion)(nil),
		(*models.PackageToBug)(nil),
		(*models.VersionToBug)(nil),
		(*models.PackageToGithubPullRequest)(nil),
		(*models.MaskToVersion)(nil),
		(*models.Package)(nil),
		(*models.CategoryPackagesInformation)(nil),
		(*models.Category)(nil),
		(*models.Version)(nil),
		(*models.Commit)(nil),
		(*models.KeywordChange)(nil),
		(*models.Useflag)(nil),
		(*models.Mask)(nil),
		(*models.OutdatedPackages)(nil),
		(*models.Project)(nil),
		(*models.MaintainerToProject)(nil),
		(*models.PkgCheckResult)(nil),
		(*models.GithubPullRequest)(nil),
		(*models.Bug)(nil),
		(*models.ReverseDependency)(nil),
		(*models.Maintainer)(nil),
		(*models.Application)(nil),
	} {
		err := DBCon.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type dbLogger struct{}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

// AfterQuery is used to log SQL queries
func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	query, err := q.FormattedQuery()
	if err == nil {
		logger.Debug.Println(string(query))
	}
	return nil
}

// Connect is used to connect to the database
// and turn on logging if desired
func Connect() {
	DBCon = pg.Connect(&pg.Options{
		User:     config.PostgresUser(),
		Password: config.PostgresPass(),
		Database: config.PostgresDb(),
		Addr:     config.PostgresHost() + ":" + config.PostgresPort(),
	})

	DBCon.AddQueryHook(dbLogger{})

	err := CreateSchema()
	if err != nil {
		logger.Error.Println("ERROR: Could not create database schema")
		logger.Error.Println(err)
		log.Fatalln(err)
	}

}

func TruncateTable[K any](primary string) {
	var val K
	var allRows []*K
	err := DBCon.Model(&allRows).Column(primary).Select()
	if err != nil {
		logger.Error.Println(err)
		return
	} else if len(allRows) == 0 {
		logger.Info.Printf("No %T to delete from the database", val)
		return
	}
	res, err := DBCon.Model(&allRows).Delete()
	if err != nil {
		logger.Error.Println(err)
		return
	}
	logger.Info.Printf("Deleted %d %T from the database", res.RowsAffected(), val)
}
