package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateUserAttribute(tx *sql.Tx, userAttribute *models.UserAttribute) error {
	return d.CommonDB.CreateUserAttribute(tx, userAttribute)
}

func (d *MySQLDB) UpdateUserAttribute(tx *sql.Tx, userAttribute *models.UserAttribute) error {
	return d.CommonDB.UpdateUserAttribute(tx, userAttribute)
}

func (d *MySQLDB) GetUserAttributeById(tx *sql.Tx, userAttributeId int64) (*models.UserAttribute, error) {
	return d.CommonDB.GetUserAttributeById(tx, userAttributeId)
}

func (d *MySQLDB) GetUserAttributesByUserId(tx *sql.Tx, userId int64) ([]models.UserAttribute, error) {
	return d.CommonDB.GetUserAttributesByUserId(tx, userId)
}

func (d *MySQLDB) DeleteUserAttribute(tx *sql.Tx, userAttributeId int64) error {
	return d.CommonDB.DeleteUserAttribute(tx, userAttributeId)
}
