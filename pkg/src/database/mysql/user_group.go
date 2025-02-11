package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/pkg/src/models"
)

func (d *MySQLDB) CreateUserGroup(tx *sql.Tx, userGroup *models.UserGroup) error {
	return d.CommonDB.CreateUserGroup(tx, userGroup)
}

func (d *MySQLDB) UpdateUserGroup(tx *sql.Tx, userGroup *models.UserGroup) error {
	return d.CommonDB.UpdateUserGroup(tx, userGroup)
}

func (d *MySQLDB) GetUserGroupById(tx *sql.Tx, userGroupId int64) (*models.UserGroup, error) {
	return d.CommonDB.GetUserGroupById(tx, userGroupId)
}

func (d *MySQLDB) GetUserGroupsByUserIds(tx *sql.Tx, userIds []int64) ([]models.UserGroup, error) {
	return d.CommonDB.GetUserGroupsByUserIds(tx, userIds)
}

func (d *MySQLDB) GetUserGroupsByUserId(tx *sql.Tx, userId int64) ([]models.UserGroup, error) {
	return d.CommonDB.GetUserGroupsByUserId(tx, userId)
}

func (d *MySQLDB) GetUserGroupByUserIdAndGroupId(tx *sql.Tx, userId, groupId int64) (*models.UserGroup, error) {
	return d.CommonDB.GetUserGroupByUserIdAndGroupId(tx, userId, groupId)
}

func (d *MySQLDB) DeleteUserGroup(tx *sql.Tx, userGroupId int64) error {
	return d.CommonDB.DeleteUserGroup(tx, userGroupId)
}
