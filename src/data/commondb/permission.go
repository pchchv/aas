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
