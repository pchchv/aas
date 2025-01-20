package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateKeyPair(tx *sql.Tx, keyPair *models.KeyPair) error {
	return d.CommonDB.CreateKeyPair(tx, keyPair)
}

func (d *MySQLDB) UpdateKeyPair(tx *sql.Tx, keyPair *models.KeyPair) error {
	return d.CommonDB.UpdateKeyPair(tx, keyPair)
}

func (d *MySQLDB) GetKeyPairById(tx *sql.Tx, keyPairId int64) (*models.KeyPair, error) {
	return d.CommonDB.GetKeyPairById(tx, keyPairId)
}

func (d *MySQLDB) GetAllSigningKeys(tx *sql.Tx) ([]models.KeyPair, error) {
	return d.CommonDB.GetAllSigningKeys(tx)
}

func (d *MySQLDB) GetCurrentSigningKey(tx *sql.Tx) (*models.KeyPair, error) {
	return d.CommonDB.GetCurrentSigningKey(tx)
}

func (d *MySQLDB) DeleteKeyPair(tx *sql.Tx, keyPairId int64) error {
	return d.CommonDB.DeleteKeyPair(tx, keyPairId)
}
