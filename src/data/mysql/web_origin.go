package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateWebOrigin(tx *sql.Tx, webOrigin *models.WebOrigin) error {
	return d.CommonDB.CreateWebOrigin(tx, webOrigin)
}

func (d *MySQLDB) GetWebOriginById(tx *sql.Tx, webOriginId int64) (*models.WebOrigin, error) {
	return d.CommonDB.GetWebOriginById(tx, webOriginId)
}

func (d *MySQLDB) GetWebOriginsByClientId(tx *sql.Tx, clientId int64) ([]models.WebOrigin, error) {
	return d.CommonDB.GetWebOriginsByClientId(tx, clientId)
}

func (d *MySQLDB) GetAllWebOrigins(tx *sql.Tx) ([]models.WebOrigin, error) {
	return d.CommonDB.GetAllWebOrigins(tx)
}

func (d *MySQLDB) DeleteWebOrigin(tx *sql.Tx, webOriginId int64) error {
	return d.CommonDB.DeleteWebOrigin(tx, webOriginId)
}
