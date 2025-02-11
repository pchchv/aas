package postgresdb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

func (d *PostgresDB) CreateResource(tx *sql.Tx, resource *models.Resource) error {
	now := time.Now().UTC()
	originalCreatedAt := resource.CreatedAt
	originalUpdatedAt := resource.UpdatedAt
	resource.CreatedAt = sql.NullTime{Time: now, Valid: true}
	resource.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	resourceStruct := sqlbuilder.NewStruct(new(models.Resource)).For(sqlbuilder.PostgreSQL)
	insertBuilder := resourceStruct.WithoutTag("pk").InsertInto("resources", resource)
	sql, args := insertBuilder.Build()
	sql += " RETURNING id"
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

func (d *PostgresDB) UpdateResource(tx *sql.Tx, resource *models.Resource) error {
	return d.CommonDB.UpdateResource(tx, resource)
}

func (d *PostgresDB) GetResourceById(tx *sql.Tx, resourceId int64) (*models.Resource, error) {
	return d.CommonDB.GetResourceById(tx, resourceId)
}

func (d *PostgresDB) GetResourceByResourceIdentifier(tx *sql.Tx, resourceIdentifier string) (*models.Resource, error) {
	return d.CommonDB.GetResourceByResourceIdentifier(tx, resourceIdentifier)
}

func (d *PostgresDB) GetResourcesByIds(tx *sql.Tx, resourceIds []int64) ([]models.Resource, error) {
	return d.CommonDB.GetResourcesByIds(tx, resourceIds)
}

func (d *PostgresDB) GetAllResources(tx *sql.Tx) ([]models.Resource, error) {
	return d.CommonDB.GetAllResources(tx)
}

func (d *PostgresDB) DeleteResource(tx *sql.Tx, resourceId int64) error {
	return d.CommonDB.DeleteResource(tx, resourceId)
}
