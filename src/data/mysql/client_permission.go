package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateClientPermission(tx *sql.Tx, clientPermission *models.ClientPermission) error {
	return d.CommonDB.CreateClientPermission(tx, clientPermission)
}

func (d *MySQLDB) UpdateClientPermission(tx *sql.Tx, clientPermission *models.ClientPermission) error {
	return d.CommonDB.UpdateClientPermission(tx, clientPermission)
}

func (d *MySQLDB) DeleteClientPermission(tx *sql.Tx, clientPermissionId int64) error {
	return d.CommonDB.DeleteClientPermission(tx, clientPermissionId)
}
