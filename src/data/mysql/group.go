package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateGroup(tx *sql.Tx, group *models.Group) error {
	return d.CommonDB.CreateGroup(tx, group)
}

func (d *MySQLDB) UpdateGroup(tx *sql.Tx, group *models.Group) error {
	return d.CommonDB.UpdateGroup(tx, group)
}
