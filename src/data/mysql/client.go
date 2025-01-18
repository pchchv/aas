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

func (d *MySQLDB) GetClientById(tx *sql.Tx, clientId int64) (*models.Client, error) {
	return d.CommonDB.GetClientById(tx, clientId)
}

func (d *MySQLDB) GetClientsByIds(tx *sql.Tx, clientIds []int64) ([]models.Client, error) {
	return d.CommonDB.GetClientsByIds(tx, clientIds)
}

func (d *MySQLDB) GetAllClients(tx *sql.Tx) ([]models.Client, error) {
	return d.CommonDB.GetAllClients(tx)
}

func (d *MySQLDB) GetClientByClientIdentifier(tx *sql.Tx, clientIdentifier string) (*models.Client, error) {
	return d.CommonDB.GetClientByClientIdentifier(tx, clientIdentifier)
}

func (d *MySQLDB) DeleteClient(tx *sql.Tx, clientId int64) error {
	return d.CommonDB.DeleteClient(tx, clientId)
}

func (d *MySQLDB) ClientLoadRedirectURIs(tx *sql.Tx, client *models.Client) error {
	return d.CommonDB.ClientLoadRedirectURIs(tx, client)
}

func (d *MySQLDB) ClientLoadWebOrigins(tx *sql.Tx, client *models.Client) error {
	return d.CommonDB.ClientLoadWebOrigins(tx, client)
}
