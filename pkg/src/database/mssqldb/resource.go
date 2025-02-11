package mssqldb

import (
	"database/sql"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

func (d *MsSQLDB) CreateResource(tx *sql.Tx, resource *models.Resource) error {
	now := time.Now().UTC()
	originalCreatedAt := resource.CreatedAt
	originalUpdatedAt := resource.UpdatedAt
	resource.CreatedAt = sql.NullTime{Time: now, Valid: true}
	resource.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	resourceStruct := sqlbuilder.NewStruct(new(models.Resource)).For(sqlbuilder.SQLServer)
	insertBuilder := resourceStruct.WithoutTag("pk").InsertInto("resources", resource)
	sql, args := insertBuilder.Build()
	parts := strings.SplitN(sql, "VALUES", 2)
	if len(parts) != 2 {
		return errors.New("unexpected SQL format from sqlbuilder")
	}

	sql = parts[0] + "OUTPUT INSERTED.id VALUES" + parts[1]
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		resource.CreatedAt = originalCreatedAt
		resource.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert resource")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&resource.Id); err != nil {
			resource.CreatedAt = originalCreatedAt
			resource.UpdatedAt = originalUpdatedAt
			return errors.Wrap(err, "unable to scan resource id")
		}
	}

	return nil
}

func (d *MsSQLDB) UpdateResource(tx *sql.Tx, resource *models.Resource) error {
	return d.CommonDB.UpdateResource(tx, resource)
}

func (d *MsSQLDB) GetResourceById(tx *sql.Tx, resourceId int64) (*models.Resource, error) {
	return d.CommonDB.GetResourceById(tx, resourceId)
}

func (d *MsSQLDB) GetResourceByResourceIdentifier(tx *sql.Tx, resourceIdentifier string) (*models.Resource, error) {
	return d.CommonDB.GetResourceByResourceIdentifier(tx, resourceIdentifier)
}

func (d *MsSQLDB) GetResourcesByIds(tx *sql.Tx, resourceIds []int64) ([]models.Resource, error) {
	return d.CommonDB.GetResourcesByIds(tx, resourceIds)
}

func (d *MsSQLDB) GetAllResources(tx *sql.Tx) ([]models.Resource, error) {
	return d.CommonDB.GetAllResources(tx)
}

func (d *MsSQLDB) DeleteResource(tx *sql.Tx, resourceId int64) error {
	return d.CommonDB.DeleteResource(tx, resourceId)
}
