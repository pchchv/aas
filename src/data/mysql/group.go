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

func (d *MySQLDB) GetGroupById(tx *sql.Tx, groupId int64) (*models.Group, error) {
	return d.CommonDB.GetGroupById(tx, groupId)
}

func (d *MySQLDB) GetGroupsByIds(tx *sql.Tx, groupIds []int64) ([]models.Group, error) {
	return d.CommonDB.GetGroupsByIds(tx, groupIds)
}

func (d *MySQLDB) GetGroupByGroupIdentifier(tx *sql.Tx, groupIdentifier string) (*models.Group, error) {
	return d.CommonDB.GetGroupByGroupIdentifier(tx, groupIdentifier)
}

func (d *MySQLDB) GetAllGroups(tx *sql.Tx) ([]models.Group, error) {
	return d.CommonDB.GetAllGroups(tx)
}

func (d *MySQLDB) GetAllGroupsPaginated(tx *sql.Tx, page int, pageSize int) ([]models.Group, int, error) {
	return d.CommonDB.GetAllGroupsPaginated(tx, page, pageSize)
}

func (d *MySQLDB) GetGroupMembersPaginated(tx *sql.Tx, groupId int64, page int, pageSize int) ([]models.User, int, error) {
	return d.CommonDB.GetGroupMembersPaginated(tx, groupId, page, pageSize)
}
