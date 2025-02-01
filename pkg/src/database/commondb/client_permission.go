package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
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

func (d *CommonDB) GetClientPermissionById(tx *sql.Tx, clientPermissionId int64) (*models.ClientPermission, error) {
	clientPermissionStruct := sqlbuilder.NewStruct(new(models.ClientPermission)).For(d.Flavor)
	selectBuilder := clientPermissionStruct.SelectFrom("clients_permissions")
	selectBuilder.Where(selectBuilder.Equal("id", clientPermissionId))
	return d.getClientPermissionCommon(tx, selectBuilder, clientPermissionStruct)
}

func (d *CommonDB) GetClientPermissionsByClientId(tx *sql.Tx, clientId int64) (clientPermissions []models.ClientPermission, err error) {
	clientPermissionStruct := sqlbuilder.NewStruct(new(models.ClientPermission)).For(d.Flavor)
	selectBuilder := clientPermissionStruct.SelectFrom("clients_permissions")
	selectBuilder.Where(selectBuilder.Equal("client_id", clientId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var clientPermission models.ClientPermission
		addr := clientPermissionStruct.Addr(&clientPermission)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan clientPermission")
		}

		clientPermissions = append(clientPermissions, clientPermission)
	}

	return clientPermissions, nil
}

func (d *CommonDB) GetClientPermissionByClientIdAndPermissionId(tx *sql.Tx, clientId, permissionId int64) (*models.ClientPermission, error) {
	clientPermissionStruct := sqlbuilder.NewStruct(new(models.ClientPermission)).For(d.Flavor)
	selectBuilder := clientPermissionStruct.SelectFrom("clients_permissions")
	selectBuilder.Where(selectBuilder.Equal("client_id", clientId))
	selectBuilder.Where(selectBuilder.Equal("permission_id", permissionId))
	return d.getClientPermissionCommon(tx, selectBuilder, clientPermissionStruct)
}

func (d *CommonDB) UpdateClientPermission(tx *sql.Tx, clientPermission *models.ClientPermission) error {
	if clientPermission.Id == 0 {
		return errors.WithStack(errors.New("can't update clientPermission with id 0"))
	}

	originalUpdatedAt := clientPermission.UpdatedAt
	clientPermission.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	clientPermissionStruct := sqlbuilder.NewStruct(new(models.ClientPermission)).For(d.Flavor)
	updateBuilder := clientPermissionStruct.WithoutTag("pk").WithoutTag("dont-update").Update("clients_permissions", clientPermission)
	updateBuilder.Where(updateBuilder.Equal("id", clientPermission.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		clientPermission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update clientPermission")
	}

	return nil
}

func (d *CommonDB) DeleteClientPermission(tx *sql.Tx, clientPermissionId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.ClientPermission)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("clients_permissions")
	deleteBuilder.Where(deleteBuilder.Equal("id", clientPermissionId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete clientPermission")
	}

	return nil
}

func (d *CommonDB) getClientPermissionCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, clientPermissionStruct *sqlbuilder.Struct) (*models.ClientPermission, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var clientPermission models.ClientPermission
	if rows.Next() {
		addr := clientPermissionStruct.Addr(&clientPermission)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan clientPermission")
		}
		return &clientPermission, nil
	}

	return nil, nil
}
