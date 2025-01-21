package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateUserConsent(tx *sql.Tx, userConsent *models.UserConsent) error {
	return d.CommonDB.CreateUserConsent(tx, userConsent)
}

func (d *MySQLDB) UpdateUserConsent(tx *sql.Tx, userConsent *models.UserConsent) error {
	return d.CommonDB.UpdateUserConsent(tx, userConsent)
}

func (d *MySQLDB) GetUserConsentById(tx *sql.Tx, userConsentId int64) (*models.UserConsent, error) {
	return d.CommonDB.GetUserConsentById(tx, userConsentId)
}

func (d *MySQLDB) GetConsentByUserIdAndClientId(tx *sql.Tx, userId int64, clientId int64) (*models.UserConsent, error) {
	return d.CommonDB.GetConsentByUserIdAndClientId(tx, userId, clientId)
}

func (d *MySQLDB) GetConsentsByUserId(tx *sql.Tx, userId int64) ([]models.UserConsent, error) {
	return d.CommonDB.GetConsentsByUserId(tx, userId)
}

func (d *MySQLDB) DeleteUserConsent(tx *sql.Tx, userConsentId int64) error {
	return d.CommonDB.DeleteUserConsent(tx, userConsentId)
}

func (d *MySQLDB) DeleteAllUserConsent(tx *sql.Tx) error {
	return d.CommonDB.DeleteAllUserConsent(tx)
}

func (d *MySQLDB) UserConsentsLoadClients(tx *sql.Tx, userConsents []models.UserConsent) error {
	return d.CommonDB.UserConsentsLoadClients(tx, userConsents)
}
