package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreatePreRegistration(tx *sql.Tx, preRegistration *models.PreRegistration) error {
	now := time.Now().UTC()
	originalCreatedAt := preRegistration.CreatedAt
	originalUpdatedAt := preRegistration.UpdatedAt
	preRegistration.CreatedAt = sql.NullTime{Time: now, Valid: true}
	preRegistration.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	preRegistrationStruct := sqlbuilder.NewStruct(new(models.PreRegistration)).For(d.Flavor)
	insertBuilder := preRegistrationStruct.WithoutTag("pk").InsertInto("pre_registrations", preRegistration)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		preRegistration.CreatedAt = originalCreatedAt
		preRegistration.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert preRegistration")
	}

	id, err := result.LastInsertId()
	if err != nil {
		preRegistration.CreatedAt = originalCreatedAt
		preRegistration.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	preRegistration.Id = id
	return nil
}
