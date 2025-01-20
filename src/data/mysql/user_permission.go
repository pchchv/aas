package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateUserPermission(tx *sql.Tx, userPermission *models.UserPermission) error {
	return d.CommonDB.CreateUserPermission(tx, userPermission)
}

func (d *MySQLDB) UpdateUserPermission(tx *sql.Tx, userPermission *models.UserPermission) error {
	return d.CommonDB.UpdateUserPermission(tx, userPermission)
}

func (d *MySQLDB) GetUserPermissionById(tx *sql.Tx, userPermissionId int64) (*models.UserPermission, error) {
	return d.CommonDB.GetUserPermissionById(tx, userPermissionId)
}

func (d *MySQLDB) GetUserPermissionsByUserIds(tx *sql.Tx, userIds []int64) ([]models.UserPermission, error) {
	return d.CommonDB.GetUserPermissionsByUserIds(tx, userIds)
}

func (d *MySQLDB) GetUserPermissionsByUserId(tx *sql.Tx, userId int64) ([]models.UserPermission, error) {
	return d.CommonDB.GetUserPermissionsByUserId(tx, userId)
}

func (d *MySQLDB) GetUserPermissionByUserIdAndPermissionId(tx *sql.Tx, userId, permissionId int64) (*models.UserPermission, error) {
	return d.CommonDB.GetUserPermissionByUserIdAndPermissionId(tx, userId, permissionId)
}

func (d *MySQLDB) GetUsersByPermissionIdPaginated(tx *sql.Tx, permissionId int64, page int, pageSize int) ([]models.User, int, error) {
	return d.CommonDB.GetUsersByPermissionIdPaginated(tx, permissionId, page, pageSize)
}

func (d *MySQLDB) DeleteUserPermission(tx *sql.Tx, userPermissionId int64) error {
	return d.CommonDB.DeleteUserPermission(tx, userPermissionId)
}
