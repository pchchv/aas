package postgresdb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

func (d *PostgresDB) CreatePermission(tx *sql.Tx, permission *models.Permission) error {
	if permission.ResourceId == 0 {
		return errors.WithStack(errors.New("can't create permission with resource_id 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := permission.CreatedAt
	originalUpdatedAt := permission.UpdatedAt
	permission.CreatedAt = sql.NullTime{Time: now, Valid: true}
	permission.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	permissionStruct := sqlbuilder.NewStruct(new(models.Permission)).For(sqlbuilder.PostgreSQL)
	insertBuilder := permissionStruct.WithoutTag("pk").InsertInto("permissions", permission)
	sql, args := insertBuilder.Build()
	sql += " RETURNING id"
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		permission.CreatedAt = originalCreatedAt
		permission.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert permission")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&permission.Id); err != nil {
			permission.CreatedAt = originalCreatedAt
			permission.UpdatedAt = originalUpdatedAt
			return errors.Wrap(err, "unable to scan permission id")
		}
	}

	return nil
}

func (d *PostgresDB) UpdatePermission(tx *sql.Tx, permission *models.Permission) error {
	return d.CommonDB.UpdatePermission(tx, permission)
}

func (d *PostgresDB) GetPermissionById(tx *sql.Tx, permissionId int64) (*models.Permission, error) {
	return d.CommonDB.GetPermissionById(tx, permissionId)
}

func (d *PostgresDB) GetPermissionsByResourceId(tx *sql.Tx, resourceId int64) ([]models.Permission, error) {
	return d.CommonDB.GetPermissionsByResourceId(tx, resourceId)
}

func (d *PostgresDB) GetPermissionsByIds(tx *sql.Tx, permissionIds []int64) ([]models.Permission, error) {
	return d.CommonDB.GetPermissionsByIds(tx, permissionIds)
}

func (d *PostgresDB) DeletePermission(tx *sql.Tx, permissionId int64) error {
	return d.CommonDB.DeletePermission(tx, permissionId)
}

func (d *PostgresDB) PermissionsLoadResources(tx *sql.Tx, permissions []models.Permission) error {
	return d.CommonDB.PermissionsLoadResources(tx, permissions)
}
