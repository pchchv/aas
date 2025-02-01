package sqlitedb

import (
	"database/sql"

	"github.com/pchchv/aas/pkg/src/models"
)

func (d *SQLiteDB) CreateGroupAttribute(tx *sql.Tx, groupAttribute *models.GroupAttribute) error {
	return d.CommonDB.CreateGroupAttribute(tx, groupAttribute)
}

func (d *SQLiteDB) UpdateGroupAttribute(tx *sql.Tx, groupAttribute *models.GroupAttribute) error {
	return d.CommonDB.UpdateGroupAttribute(tx, groupAttribute)
}

func (d *SQLiteDB) GetGroupAttributeById(tx *sql.Tx, groupAttributeId int64) (*models.GroupAttribute, error) {
	return d.CommonDB.GetGroupAttributeById(tx, groupAttributeId)
}

func (d *SQLiteDB) GetGroupAttributesByGroupIds(tx *sql.Tx, groupIds []int64) ([]models.GroupAttribute, error) {
	return d.CommonDB.GetGroupAttributesByGroupIds(tx, groupIds)
}

func (d *SQLiteDB) GetGroupAttributesByGroupId(tx *sql.Tx, groupId int64) ([]models.GroupAttribute, error) {
	return d.CommonDB.GetGroupAttributesByGroupId(tx, groupId)
}

func (d *SQLiteDB) DeleteGroupAttribute(tx *sql.Tx, groupAttributeId int64) error {
	return d.CommonDB.DeleteGroupAttribute(tx, groupAttributeId)
}
