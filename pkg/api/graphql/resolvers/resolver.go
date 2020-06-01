package resolvers

// THIS WILL NOT BE AUTOMATICALLY UPDATED WITH SCHEMA CHANGES.

import (
	"context"
	"errors"
	"github.com/go-pg/pg/v9/orm"
	"soko/pkg/api/graphql/generated"
	"soko/pkg/app/handler/packages"
	"soko/pkg/database"
	"soko/pkg/models"
	"time"
)

type Resolver struct{}

func (r *queryResolver) Category(ctx context.Context, name *string, description *string) (*models.Category, error) {
	categories, err := r.Categories(ctx, name, description)
	if err != nil || len(categories) != 1 {
		return &models.Category{}, errors.New("your parameters do not uniquely match a category")
	}
	return categories[0], nil
}

func (r *queryResolver) Categories(ctx context.Context, name *string, description *string) ([]*models.Category, error) {
	var categories []*models.Category
	query := database.DBCon.Model(&categories)
	params := map[string]*string{
		"name":        name,
		"description": description,
	}
	query = addStringParams(query, params)
	err := query.Relation("Packages").Select()
	if err != nil {
		return []*models.Category{}, errors.New("an error occurred while searching for the categories")
	}
	return categories, nil
}

func (r *queryResolver) Commit(ctx context.Context, id *string, precedingCommits *int, authorName *string, authorEmail *string, authorDate *time.Time, committerName *string, committerEmail *string, committerDate *time.Time, message *string) (*models.Commit, error) {
	commits, err := r.Commits(ctx, id, precedingCommits, authorName, authorEmail, authorDate, committerName, committerName, committerDate, message)
	if err != nil || len(commits) != 1 {
		return &models.Commit{}, errors.New("your parameters do not uniquely match a commit")
	}
	return commits[0], nil
}

func (r *queryResolver) Commits(ctx context.Context, id *string, precedingCommits *int, authorName *string, authorEmail *string, authorDate *time.Time, committerName *string, committerEmail *string, committerDate *time.Time, message *string) ([]*models.Commit, error) {
	var commits []*models.Commit
	query := database.DBCon.Model(&commits)
	stringParams := map[string]*string{
		"id":              id,
		"author_name":     authorName,
		"author_email":    authorEmail,
		"committer_name":  committerName,
		"committer_email": committerEmail,
		"message":         message,
	}
	intParams := map[string]*int{
		"preceding_commits": precedingCommits,
	}
	timeParams := map[string]*time.Time{
		"author_date":    authorDate,
		"committer_date": committerDate,
	}
	query = addStringParams(query, stringParams)
	query = addIntParams(query, intParams)
	query = addTimeParams(query, timeParams)
	err := query.Relation("ChangedPackages").Relation("ChangedVersions").Relation("KeywordChanges").Select()
	if err != nil {
		return []*models.Commit{}, errors.New("an error occurred while searching for the commits")
	}
	return commits, nil
}

func (r *queryResolver) Mask(ctx context.Context, versions *string, author *string, authorEmail *string, date *time.Time, reason *string) (*models.Mask, error) {
	masks, err := r.Masks(ctx, versions, author, authorEmail, date, reason)
	if err != nil || len(masks) != 1 {
		return &models.Mask{}, errors.New("your parameters do not uniquely match a mask")
	}
	return masks[0], nil
}

func (r *queryResolver) Masks(ctx context.Context, versions *string, author *string, authorEmail *string, date *time.Time, reason *string) ([]*models.Mask, error) {
	var masks []*models.Mask
	query := database.DBCon.Model(&masks)
	stringParams := map[string]*string{
		"versions":     versions,
		"author":       author,
		"author_email": authorEmail,
		"reason":       reason,
	}
	timeParams := map[string]*time.Time{
		"date": date,
	}
	query = addStringParams(query, stringParams)
	query = addTimeParams(query, timeParams)
	err := query.Select()
	if err != nil {
		return []*models.Mask{}, errors.New("an error occurred while searching for the masks")
	}
	return masks, nil
}

func (r *queryResolver) OutdatedPackage(ctx context.Context, atom *string, gentooVersion *string, newestVersion *string) (*models.OutdatedPackages, error) {
	outdatedPackages, err := r.OutdatedPackages(ctx, atom, gentooVersion, newestVersion)
	if err != nil || len(outdatedPackages) != 1 {
		return &models.OutdatedPackages{}, errors.New("your parameters do not uniquely match a outdated Version")
	}
	return outdatedPackages[0], nil
}

func (r *queryResolver) OutdatedPackages(ctx context.Context, atom *string, gentooVersion *string, newestVersion *string) ([]*models.OutdatedPackages, error) {
	var outdatedPackages []*models.OutdatedPackages
	query := database.DBCon.Model(&outdatedPackages)
	params := map[string]*string{
		"atom":           atom,
		"gentoo_version": gentooVersion,
		"newest_version": newestVersion,
	}
	query = addStringParams(query, params)
	err := query.Select()
	if err != nil {
		return []*models.OutdatedPackages{}, errors.New("an error occurred while searching for the outdated packages")
	}
	return outdatedPackages, nil
}

func (r *queryResolver) PkgCheckResult(ctx context.Context, atom *string, category *string, packageArg *string, version *string, cpv *string, class *string, message *string) (*models.PkgCheckResult, error) {
	pkgCheckResults, err := r.PkgCheckResults(ctx, atom, category, packageArg, version, cpv, class, message)
	if err != nil || len(pkgCheckResults) != 1 {
		return &models.PkgCheckResult{}, errors.New("your parameters do not uniquely match a pkgcheck result")
	}
	return pkgCheckResults[0], nil
}

func (r *queryResolver) PkgCheckResults(ctx context.Context, atom *string, category *string, packageArg *string, version *string, cpv *string, class *string, message *string) ([]*models.PkgCheckResult, error) {
	var pkgCheckResults []*models.PkgCheckResult
	query := database.DBCon.Model(&pkgCheckResults)
	params := map[string]*string{
		"atom":     atom,
		"category": category,
		"package":  packageArg,
		"version":  version,
		"cpv":      cpv,
		"class":    class,
		"message":  message,
	}
	query = addStringParams(query, params)
	err := query.Select()
	if err != nil {
		return []*models.PkgCheckResult{}, errors.New("an error occurred while searching for the pkgcheck results")
	}
	return pkgCheckResults, nil
}

func (r *queryResolver) Package(ctx context.Context, atom *string, category *string, name *string, longdescription *string, precedingCommits *int) (*models.Package, error) {
	gpackages, err := r.Packages(ctx, atom, category, name, longdescription, precedingCommits)
	if err != nil || len(gpackages) != 1 {
		return &models.Package{}, errors.New("your parameters do not uniquely match a package")
	}
	return gpackages[0], nil
}

func (r *queryResolver) Packages(ctx context.Context, atom *string, category *string, name *string, longdescription *string, precedingCommits *int) ([]*models.Package, error) {
	var gpackages []*models.Package
	query := database.DBCon.Model(&gpackages)
	stringParams := map[string]*string{
		"atom":            atom,
		"category":        category,
		"name":            name,
		"longdescription": longdescription,
	}
	intParams := map[string]*int{
		"preceding_commits": precedingCommits,
	}
	query = addStringParams(query, stringParams)
	query = addIntParams(query, intParams)
	err := query.Relation("Commits").Relation("Versions").Select()
	if err != nil {
		return []*models.Package{}, errors.New("an error occurred while searching for the packages")
	}
	return gpackages, nil
}

func (r *queryResolver) Version(ctx context.Context, id *string, category *string, packageArg *string, atom *string, version *string, slot *string, subslot *string, eapi *string, keywords *string, useflags *string, restricts *string, properties *string, homepage *string, license *string, description *string) (*models.Version, error) {
	versions, err := r.Versions(ctx, id, category, packageArg, atom, version, slot, subslot, eapi, keywords, useflags, restricts, properties, homepage, license, description)
	if err != nil || len(versions) != 1 {
		return &models.Version{}, errors.New("your parameters do not uniquely match a version")
	}
	return versions[0], nil
}

func (r *queryResolver) Versions(ctx context.Context, id *string, category *string, packageArg *string, atom *string, version *string, slot *string, subslot *string, eapi *string, keywords *string, useflags *string, restricts *string, properties *string, homepage *string, license *string, description *string) ([]*models.Version, error) {
	var versions []*models.Version
	query := database.DBCon.Model(&versions)
	params := map[string]*string{
		"id":          id,
		"category":    category,
		"atom":        atom,
		"package":     packageArg,
		"version":     version,
		"slot":        slot,
		"subslot":     subslot,
		"eapi":        eapi,
		"keywords":    keywords,
		"useflags":    useflags,
		"restricts":   restricts,
		"properties":  properties,
		"homepage":    homepage,
		"license":     license,
		"description": description,
	}
	query = addStringParams(query, params)
	err := query.Relation("Commits").Relation("Masks").Select()
	if err != nil {
		return []*models.Version{}, errors.New("an error occurred while searching for the versions")
	}
	return versions, nil
}

func (r *queryResolver) Useflag(ctx context.Context, id *string, name *string, scope *string, description *string, useExpand *string, packageArg *string) (*models.Useflag, error) {
	useflags, err := r.Useflags(ctx, id, name, scope, description, useExpand, packageArg)
	if err != nil || len(useflags) != 1 {
		return &models.Useflag{}, errors.New("your parameters do not uniquely match a useflag")
	}
	return useflags[0], nil
}

func (r *queryResolver) Useflags(ctx context.Context, id *string, name *string, scope *string, description *string, useExpand *string, packageArg *string) ([]*models.Useflag, error) {
	var useflags []*models.Useflag
	query := database.DBCon.Model(&useflags)
	params := map[string]*string{
		"id":          id,
		"name":        name,
		"scope":       scope,
		"description": description,
		"use_expand":  useExpand,
		"package":     packageArg,
	}
	query = addStringParams(query, params)
	err := query.Select()
	if err != nil {
		return []*models.Useflag{}, errors.New("an error occurred while searching for the useflags")
	}
	return useflags, nil
}

func (r *queryResolver) AddedPackages(ctx context.Context, limit *int) ([]*models.Package, error) {
	n := getLimit(limit, 25)
	return packages.GetAddedPackages(n), nil
}

func (r *queryResolver) UpdatedVersions(ctx context.Context, limit *int) ([]*models.Version, error) {
	n := getLimit(limit, 25)
	return packages.GetUpdatedVersions(n), nil
}

func (r *queryResolver) StabilizedVersions(ctx context.Context, limit *int, arch *string) ([]*models.Version, error) {
	n := getLimit(limit, 25)
	return packages.GetStabilizedVersions(n), nil
}

func (r *queryResolver) KeywordedVersions(ctx context.Context, limit *int, arch *string) ([]*models.Version, error) {
	n := getLimit(limit, 25)
	return packages.GetKeywordedVersions(n), nil
}

// utility functions

func getLimit(limit *int, defaultLimit int) int {
	var n int
	if limit != nil {
		n = *limit
	} else {
		n = defaultLimit
	}
	return n
}

func addStringParams(query *orm.Query, params map[string]*string) *orm.Query {
	for key, value := range params {
		if value != nil {
			query = query.Where(key+" = ? ", *value)
		}
	}
	return query
}

func addIntParams(query *orm.Query, params map[string]*int) *orm.Query {
	for key, value := range params {
		if value != nil {
			query = query.Where(key+" = ? ", *value)
		}
	}
	return query
}

func addTimeParams(query *orm.Query, params map[string]*time.Time) *orm.Query {
	for key, value := range params {
		if value != nil {
			query = query.Where(key+" = ? ", *value)
		}
	}
	return query
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
