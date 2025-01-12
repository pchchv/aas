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

func (d *CommonDB) GetPermissionById(tx *sql.Tx, permissionId int64) (*models.Permission, error) {
	permissionStruct := sqlbuilder.NewStruct(new(models.Permission)).For(d.Flavor)
	selectBuilder := permissionStruct.SelectFrom("permissions")
	selectBuilder.Where(selectBuilder.Equal("id", permissionId))
	return d.getPermissionCommon(tx, selectBuilder, permissionStruct)
}

func (d *CommonDB) UpdatePermission(tx *sql.Tx, permission *models.Permission) error {
	if permission.Id == 0 {
		return errors.WithStack(errors.New("can't update permission with id 0"))
	}

	originalUpdatedAt := permission.UpdatedAt
	permission.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	permissionStruct := sqlbuilder.NewStruct(new(models.Permission)).For(d.Flavor)
	updateBuilder := permissionStruct.WithoutTag("pk").WithoutTag("dont-update").Update("permissions", permission)
	updateBuilder.Where(updateBuilder.Equal("id", permission.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		permission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update permission")
	}

	return nil
}

func (d *CommonDB) DeletePermission(tx *sql.Tx, permissionId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.Permission)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("permissions")
	deleteBuilder.Where(deleteBuilder.Equal("id", permissionId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete permission")
	}

	return nil
}

func (d *CommonDB) PermissionsLoadResources(tx *sql.Tx, permissions []models.Permission) error {
	if permissions == nil {
		return nil
	}

	resourceIds := make([]int64, 0, len(permissions))
	for _, permission := range permissions {
		resourceIds = append(resourceIds, permission.ResourceId)
	}

	resources, err := d.GetResourcesByIds(tx, resourceIds)
	if err != nil {
		return errors.Wrap(err, "unable to get resources for permissions")
	}

	resourceMap := make(map[int64]models.Resource, len(resources))
	for _, resource := range resources {
		resourceMap[resource.Id] = resource
	}

	for i := range permissions {
		permissions[i].Resource = resourceMap[permissions[i].ResourceId]
	}

	return nil
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
