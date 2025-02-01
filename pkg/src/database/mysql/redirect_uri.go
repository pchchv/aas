package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/pkg/src/models"
)

func (d *MySQLDB) CreateRedirectURI(tx *sql.Tx, redirectURI *models.RedirectURI) error {
	return d.CommonDB.CreateRedirectURI(tx, redirectURI)
}

func (d *MySQLDB) GetRedirectURIById(tx *sql.Tx, redirectURIId int64) (*models.RedirectURI, error) {
	return d.CommonDB.GetRedirectURIById(tx, redirectURIId)
}

func (d *MySQLDB) GetRedirectURIsByClientId(tx *sql.Tx, clientId int64) ([]models.RedirectURI, error) {
	return d.CommonDB.GetRedirectURIsByClientId(tx, clientId)
}

func (d *MySQLDB) DeleteRedirectURI(tx *sql.Tx, redirectURIId int64) error {
	return d.CommonDB.DeleteRedirectURI(tx, redirectURIId)
}
