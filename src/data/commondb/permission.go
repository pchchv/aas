package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreatePermission(tx *sql.Tx, permission *models.Permission) error {
	if permission.ResourceId == 0 {
		return errors.WithStack(errors.New("can't create permission with resource_id 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := permission.CreatedAt
	originalUpdatedAt := permission.UpdatedAt
	permission.CreatedAt = sql.NullTime{Time: now, Valid: true}
	permission.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	permissionStruct := sqlbuilder.NewStruct(new(models.Permission)).For(d.Flavor)
	insertBuilder := permissionStruct.WithoutTag("pk").InsertInto("permissions", permission)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		permission.CreatedAt = originalCreatedAt
		permission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert permission")
	}

	id, err := result.LastInsertId()
	if err != nil {
		permission.CreatedAt = originalCreatedAt
		permission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	permission.Id = id
	return nil
}

func (d *CommonDB) GetPermissionsByResourceId(tx *sql.Tx, resourceId int64) (permissions []models.Permission, err error) {
	permissionStruct := sqlbuilder.NewStruct(new(models.Permission)).For(d.Flavor)
	selectBuilder := permissionStruct.SelectFrom("permissions")
	selectBuilder.Where(selectBuilder.Equal("resource_id", resourceId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var permission models.Permission
		addr := permissionStruct.Addr(&permission)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan permission")
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (d *CommonDB) GetPermissionsByIds(tx *sql.Tx, permissionIds []int64) (permissions []models.Permission, err error) {
	if len(permissionIds) == 0 {
		return
	}

	permissionStruct := sqlbuilder.NewStruct(new(models.Permission)).For(d.Flavor)
	selectBuilder := permissionStruct.SelectFrom("permissions")
	selectBuilder.Where(selectBuilder.In("id", sqlbuilder.Flatten(permissionIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var permission models.Permission
		addr := permissionStruct.Addr(&permission)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan permission")
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (d *CommonDB) getPermissionCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, permissionStruct *sqlbuilder.Struct) (*models.Permission, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var permission models.Permission
	if rows.Next() {
		addr := permissionStruct.Addr(&permission)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan permission")
		}

		return &permission, nil
	}

	return nil, nil
}
