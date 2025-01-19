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

func (d *MySQLDB) GetCodeById(tx *sql.Tx, codeId int64) (*models.Code, error) {
	return d.CommonDB.GetCodeById(tx, codeId)
}

func (d *MySQLDB) GetCodeByCodeHash(tx *sql.Tx, codeHash string, used bool) (*models.Code, error) {
	return d.CommonDB.GetCodeByCodeHash(tx, codeHash, used)
}
