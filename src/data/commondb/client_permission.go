package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateClientPermission(tx *sql.Tx, clientPermission *models.ClientPermission) error {
	if clientPermission.ClientId == 0 {
		return errors.WithStack(errors.New("can't create clientPermission with client_id 0"))
	}

	if clientPermission.PermissionId == 0 {
		return errors.WithStack(errors.New("can't create clientPermission with permission_id 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := clientPermission.CreatedAt
	originalUpdatedAt := clientPermission.UpdatedAt
	clientPermission.CreatedAt = sql.NullTime{Time: now, Valid: true}
	clientPermission.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	clientPermissionStruct := sqlbuilder.NewStruct(new(models.ClientPermission)).For(d.Flavor)
	insertBuilder := clientPermissionStruct.WithoutTag("pk").InsertInto("clients_permissions", clientPermission)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		clientPermission.CreatedAt = originalCreatedAt
		clientPermission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert clientPermission")
	}

	if id, err := result.LastInsertId(); err != nil {
		clientPermission.CreatedAt = originalCreatedAt
		clientPermission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	} else {
		clientPermission.Id = id
	}

	return nil
}
