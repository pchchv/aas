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
