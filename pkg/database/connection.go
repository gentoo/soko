// SPDX-License-Identifier: GPL-2.0-only
// Contains utility functions around the database

package database

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"

	"soko/pkg/config"
	"soko/pkg/models"
)

// DBCon is the connection handle
// for the database
var (
	DBCon *pg.DB
)

// CreateSchema creates the tables in the database
// in case they don't already exist
func CreateSchema() error {
	for _, model := range []any{
		(*models.CommitToPackage)(nil),
		(*models.CommitToVersion)(nil),
		(*models.PackageToBug)(nil),
		(*models.VersionToBug)(nil),
		(*models.PackageToGithubPullRequest)(nil),
		(*models.MaskToVersion)(nil),
		(*models.DeprecatedToVersion)(nil),
		(*models.Package)(nil),
		(*models.PkgMove)(nil),
		(*models.CategoryPackagesInformation)(nil),
		(*models.Category)(nil),
		(*models.Version)(nil),
		(*models.Commit)(nil),
		(*models.KeywordChange)(nil),
		(*models.Useflag)(nil),
		(*models.Mask)(nil),
		(*models.DeprecatedPackage)(nil),
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
			tableName := string(DBCon.Model(model).TableModel().Table().TypeName)
			slog.Error("Failed creating table", slog.String("table", tableName), slog.Any("err", err))
			return err
		}
	}
	_, err := DBCon.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm")
	if err != nil {
		slog.Error("Failed creating extension 'pg_trgm'", slog.Any("err", err))
		return err
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
		slog.Debug(string(query), slog.Duration("duration", time.Since(q.StartTime))) //nolint: sloglint
	}
	return nil
}

// Connect is used to connect to the database
// and turn on logging if desired
func Connect() {
	DBCon = pg.Connect(&pg.Options{
		User:        config.PostgresUser(),
		Password:    config.PostgresPass(),
		Database:    config.PostgresDb(),
		Addr:        config.PostgresHost() + ":" + config.PostgresPort(),
		DialTimeout: 10 * time.Second,
	})

	if !config.Quiet() {
		DBCon.AddQueryHook(dbLogger{})
	}

	if err := CreateSchema(); err != nil {
		slog.Error("Failed creating database schema", slog.Any("err", err))
		os.Exit(1)
	}
}

func TruncateTable(model any) {
	query := DBCon.Model(model)
	tableName := string(query.TableModel().Table().TypeName)
	_, err := query.Exec("TRUNCATE TABLE ?TableName")
	if err != nil {
		slog.Error("Failed truncating table", slog.String("table", tableName), slog.Any("err", err))
	} else {
		slog.Info("Truncated table", slog.String("table", tableName))
	}
}
