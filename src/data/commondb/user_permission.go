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
