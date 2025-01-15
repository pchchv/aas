package commondb

import (
	"database/sql"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) GetUserPermissionsByUserIds(tx *sql.Tx, userIds []int64) (userPermissions []models.UserPermission, err error) {
	if len(userIds) == 0 {
		return nil, nil
	}

	userPermissionStruct := sqlbuilder.NewStruct(new(models.UserPermission)).For(d.Flavor)
	selectBuilder := userPermissionStruct.SelectFrom("users_permissions")
	selectBuilder.Where(selectBuilder.In("user_id", sqlbuilder.Flatten(userIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userPermission models.UserPermission
		addr := userPermissionStruct.Addr(&userPermission)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userPermission")
		}
		userPermissions = append(userPermissions, userPermission)
	}

	return
}

func (d *CommonDB) GetUserPermissionsByUserId(tx *sql.Tx, userId int64) (userPermissions []models.UserPermission, err error) {
	userPermissionStruct := sqlbuilder.NewStruct(new(models.UserPermission)).For(d.Flavor)
	selectBuilder := userPermissionStruct.SelectFrom("users_permissions")
	selectBuilder.Where(selectBuilder.Equal("user_id", userId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userPermission models.UserPermission
		addr := userPermissionStruct.Addr(&userPermission)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan userPermission")
		}
		userPermissions = append(userPermissions, userPermission)
	}

	return
}

func (d *CommonDB) GetUsersByPermissionIdPaginated(tx *sql.Tx, permissionId int64, page int, pageSize int) (users []models.User, total int, err error) {
	if permissionId <= 0 {
		return nil, 0, errors.WithStack(errors.New("permissionId must be greater than 0"))
	}

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	selectBuilder := userStruct.SelectFrom("users")
	selectBuilder.JoinWithOption(sqlbuilder.InnerJoin, "users_permissions", "users.id = users_permissions.user_id")
	selectBuilder.Where(selectBuilder.Equal("users_permissions.permission_id", permissionId))
	selectBuilder.OrderBy("users.given_name").Asc()
	selectBuilder.Offset((page - 1) * pageSize)
	selectBuilder.Limit(pageSize)
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		addr := userStruct.Addr(&user)
		if err = rows.Scan(addr...); err != nil {
			return nil, 0, errors.Wrap(err, "unable to scan user")
		}
		users = append(users, user)
	}

	selectBuilder = d.Flavor.NewSelectBuilder()
	selectBuilder.Select("count(*)").From("users")
	selectBuilder.JoinWithOption(sqlbuilder.InnerJoin, "users_permissions", "users.id = users_permissions.user_id")
	selectBuilder.Where(selectBuilder.Equal("users_permissions.permission_id", permissionId))
	sql, args = selectBuilder.Build()
	rows2, err := d.QuerySql(nil, sql, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to query database")
	}
	defer rows2.Close()

	if rows2.Next() {
		if err = rows2.Scan(&total); err != nil {
			return nil, 0, errors.Wrap(err, "unable to scan total")
		}
	}

	return
}

func (d *CommonDB) getUserPermissionCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, userPermissionStruct *sqlbuilder.Struct) (*models.UserPermission, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var userPermission models.UserPermission
	if rows.Next() {
		addr := userPermissionStruct.Addr(&userPermission)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userPermission")
		}
		return &userPermission, nil
	}

	return nil, nil
}
