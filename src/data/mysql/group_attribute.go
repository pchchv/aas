package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) GetGroupAttributeById(tx *sql.Tx, groupAttributeId int64) (*models.GroupAttribute, error) {
	return d.CommonDB.GetGroupAttributeById(tx, groupAttributeId)
}

func (d *MySQLDB) GetGroupAttributesByGroupIds(tx *sql.Tx, groupIds []int64) ([]models.GroupAttribute, error) {
	return d.CommonDB.GetGroupAttributesByGroupIds(tx, groupIds)
}

func (d *MySQLDB) GetGroupAttributesByGroupId(tx *sql.Tx, groupId int64) ([]models.GroupAttribute, error) {
	return d.CommonDB.GetGroupAttributesByGroupId(tx, groupId)
}
