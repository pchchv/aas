package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateClient(tx *sql.Tx, client *models.Client) error {
	return d.CommonDB.CreateClient(tx, client)
}

func (d *MySQLDB) UpdateClient(tx *sql.Tx, client *models.Client) error {
	return d.CommonDB.UpdateClient(tx, client)
}
