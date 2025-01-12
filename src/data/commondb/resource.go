package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateResource(tx *sql.Tx, resource *models.Resource) error {
	now := time.Now().UTC()
	originalCreatedAt := resource.CreatedAt
	originalUpdatedAt := resource.UpdatedAt
	resource.CreatedAt = sql.NullTime{Time: now, Valid: true}
	resource.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	resourceStruct := sqlbuilder.NewStruct(new(models.Resource)).For(d.Flavor)
	insertBuilder := resourceStruct.WithoutTag("pk").InsertInto("resources", resource)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		resource.CreatedAt = originalCreatedAt
		resource.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert resource")
	}

	id, err := result.LastInsertId()
	if err != nil {
		resource.CreatedAt = originalCreatedAt
		resource.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	resource.Id = id
	return nil
}

func (d *CommonDB) GetResourcesByIds(tx *sql.Tx, resourceIds []int64) (resources []models.Resource, err error) {
	if len(resourceIds) == 0 {
		return nil, nil
	}

	resourceStruct := sqlbuilder.NewStruct(new(models.Resource)).For(d.Flavor)
	selectBuilder := resourceStruct.SelectFrom("resources")
	selectBuilder.Where(selectBuilder.In("id", sqlbuilder.Flatten(resourceIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var resource models.Resource
		addr := resourceStruct.Addr(&resource)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan resource")
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (d *CommonDB) GetAllResources(tx *sql.Tx) (resources []models.Resource, err error) {
	resourceStruct := sqlbuilder.NewStruct(new(models.Resource)).For(d.Flavor)
	selectBuilder := resourceStruct.SelectFrom("resources")
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var resource models.Resource
		addr := resourceStruct.Addr(&resource)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan resource")
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (d *CommonDB) getResourceCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, resourceStruct *sqlbuilder.Struct) (*models.Resource, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var resource models.Resource
	if rows.Next() {
		addr := resourceStruct.Addr(&resource)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan resource")
		}
		return &resource, nil
	}

	return nil, nil
}
