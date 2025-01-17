package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateGroupPermission(tx *sql.Tx, groupPermission *models.GroupPermission) error {
	if groupPermission.GroupId == 0 {
		return errors.WithStack(errors.New("can't create groupPermission with group_id 0"))
	}

	if groupPermission.PermissionId == 0 {
		return errors.WithStack(errors.New("can't create groupPermission with permission_id 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := groupPermission.CreatedAt
	originalUpdatedAt := groupPermission.UpdatedAt
	groupPermission.CreatedAt = sql.NullTime{Time: now, Valid: true}
	groupPermission.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	groupPermissionStruct := sqlbuilder.NewStruct(new(models.GroupPermission)).For(d.Flavor)
	insertBuilder := groupPermissionStruct.WithoutTag("pk").InsertInto("groups_permissions", groupPermission)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		groupPermission.CreatedAt = originalCreatedAt
		groupPermission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert groupPermission")
	}

	id, err := result.LastInsertId()
	if err != nil {
		groupPermission.CreatedAt = originalCreatedAt
		groupPermission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	groupPermission.Id = id
	return nil
}

func (d *CommonDB) GetGroupPermissionsByGroupId(tx *sql.Tx, groupId int64) (groupPermissions []models.GroupPermission, err error) {
	groupPermissionStruct := sqlbuilder.NewStruct(new(models.GroupPermission)).For(d.Flavor)
	selectBuilder := groupPermissionStruct.SelectFrom("groups_permissions")
	selectBuilder.Where(selectBuilder.Equal("group_id", groupId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var groupPermission models.GroupPermission
		addr := groupPermissionStruct.Addr(&groupPermission)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan groupPermission")
		}
		groupPermissions = append(groupPermissions, groupPermission)
	}

	return
}

func (d *CommonDB) GetGroupPermissionsByGroupIds(tx *sql.Tx, groupIds []int64) (groupPermissions []models.GroupPermission, err error) {
	if len(groupIds) == 0 {
		return nil, nil
	}

	groupPermissionStruct := sqlbuilder.NewStruct(new(models.GroupPermission)).For(d.Flavor)
	selectBuilder := groupPermissionStruct.SelectFrom("groups_permissions")
	selectBuilder.Where(selectBuilder.In("group_id", sqlbuilder.Flatten(groupIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var groupPermission models.GroupPermission
		addr := groupPermissionStruct.Addr(&groupPermission)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan groupPermission")
		}
		groupPermissions = append(groupPermissions, groupPermission)
	}

	return
}

func (d *CommonDB) getGroupPermissionCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, groupPermissionStruct *sqlbuilder.Struct) (*models.GroupPermission, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var groupPermission models.GroupPermission
	if rows.Next() {
		addr := groupPermissionStruct.Addr(&groupPermission)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan groupPermission")
		}
		return &groupPermission, nil
	}

	return nil, nil
}
