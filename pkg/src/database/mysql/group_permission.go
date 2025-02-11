package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/pkg/src/models"
)

func (d *MySQLDB) CreateGroupPermission(tx *sql.Tx, groupPermission *models.GroupPermission) error {
	return d.CommonDB.CreateGroupPermission(tx, groupPermission)
}

func (d *MySQLDB) UpdateGroupPermission(tx *sql.Tx, groupPermission *models.GroupPermission) error {
	return d.CommonDB.UpdateGroupPermission(tx, groupPermission)
}

func (d *MySQLDB) GetGroupPermissionsByGroupId(tx *sql.Tx, groupId int64) ([]models.GroupPermission, error) {
	return d.CommonDB.GetGroupPermissionsByGroupId(tx, groupId)
}

func (d *MySQLDB) GetGroupPermissionsByGroupIds(tx *sql.Tx, groupIds []int64) ([]models.GroupPermission, error) {
	return d.CommonDB.GetGroupPermissionsByGroupIds(tx, groupIds)
}

func (d *MySQLDB) GetGroupPermissionById(tx *sql.Tx, groupPermissionId int64) (*models.GroupPermission, error) {
	return d.CommonDB.GetGroupPermissionById(tx, groupPermissionId)
}

func (d *MySQLDB) GetGroupPermissionByGroupIdAndPermissionId(tx *sql.Tx, groupId, permissionId int64) (*models.GroupPermission, error) {
	return d.CommonDB.GetGroupPermissionByGroupIdAndPermissionId(tx, groupId, permissionId)
}

func (d *MySQLDB) DeleteGroupPermission(tx *sql.Tx, groupPermissionId int64) error {
	return d.CommonDB.DeleteGroupPermission(tx, groupPermissionId)
}
