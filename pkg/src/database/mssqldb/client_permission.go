package mssqldb

import (
	"database/sql"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

func (d *MsSQLDB) CreateClientPermission(tx *sql.Tx, clientPermission *models.ClientPermission) error {
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
	clientPermissionStruct := sqlbuilder.NewStruct(new(models.ClientPermission)).For(sqlbuilder.SQLServer)
	insertBuilder := clientPermissionStruct.WithoutTag("pk").InsertInto("clients_permissions", clientPermission)
	sql, args := insertBuilder.Build()
	parts := strings.SplitN(sql, "VALUES", 2)
	if len(parts) != 2 {
		return errors.New("unexpected SQL format from sqlbuilder")
	}

	sql = parts[0] + "OUTPUT INSERTED.id VALUES" + parts[1]
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		clientPermission.CreatedAt = originalCreatedAt
		clientPermission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert clientPermission")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&clientPermission.Id); err != nil {
			clientPermission.CreatedAt = originalCreatedAt
			clientPermission.UpdatedAt = originalUpdatedAt
			return errors.Wrap(err, "unable to scan clientPermission id")
		}
	}

	return nil
}

func (d *MsSQLDB) UpdateClientPermission(tx *sql.Tx, clientPermission *models.ClientPermission) error {
	return d.CommonDB.UpdateClientPermission(tx, clientPermission)
}

func (d *MsSQLDB) GetClientPermissionById(tx *sql.Tx, clientPermissionId int64) (*models.ClientPermission, error) {
	return d.CommonDB.GetClientPermissionById(tx, clientPermissionId)
}

func (d *MsSQLDB) GetClientPermissionByClientIdAndPermissionId(tx *sql.Tx, clientId, permissionId int64) (*models.ClientPermission, error) {
	return d.CommonDB.GetClientPermissionByClientIdAndPermissionId(tx, clientId, permissionId)
}

func (d *MsSQLDB) GetClientPermissionsByClientId(tx *sql.Tx, clientId int64) ([]models.ClientPermission, error) {
	return d.CommonDB.GetClientPermissionsByClientId(tx, clientId)
}

func (d *MsSQLDB) DeleteClientPermission(tx *sql.Tx, clientPermissionId int64) error {
	return d.CommonDB.DeleteClientPermission(tx, clientPermissionId)
}
