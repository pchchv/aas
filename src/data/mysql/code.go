package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateCode(tx *sql.Tx, code *models.Code) error {
	return d.CommonDB.CreateCode(tx, code)
}

func (d *MySQLDB) UpdateCode(tx *sql.Tx, code *models.Code) error {
	return d.CommonDB.UpdateCode(tx, code)
}
