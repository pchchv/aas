package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateResource(tx *sql.Tx, resource *models.Resource) error {
	now := time.Now().UTC()
	originalCreatedAt := resource.CreatedAt
	originalUpdatedAt := resource.UpdatedAt
	resource.CreatedAt = sql.NullTime{Time: now, Valid: true}
	resource.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	resourceStruct := sqlbuilder.NewStruct(new(models.Resource)).For(d.Flavor)
	insertBuilder := resourceStruct.WithoutTag("pk").InsertInto("resources", resource)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		resource.CreatedAt = originalCreatedAt
		resource.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert resource")
	}

	id, err := result.LastInsertId()
	if err != nil {
		resource.CreatedAt = originalCreatedAt
		resource.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	resource.Id = id
	return nil
}
