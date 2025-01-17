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

func (d *CommonDB) GetPreRegistrationByEmail(tx *sql.Tx, email string) (*models.PreRegistration, error) {
	preRegistrationStruct := sqlbuilder.NewStruct(new(models.PreRegistration)).For(d.Flavor)
	selectBuilder := preRegistrationStruct.SelectFrom("pre_registrations")
	selectBuilder.Where(selectBuilder.Equal("email", email))
	return d.getPreRegistrationCommon(tx, selectBuilder, preRegistrationStruct)
}

func (d *CommonDB) GetPreRegistrationById(tx *sql.Tx, preRegistrationId int64) (*models.PreRegistration, error) {
	preRegistrationStruct := sqlbuilder.NewStruct(new(models.PreRegistration)).For(d.Flavor)
	selectBuilder := preRegistrationStruct.SelectFrom("pre_registrations")
	selectBuilder.Where(selectBuilder.Equal("id", preRegistrationId))
	return d.getPreRegistrationCommon(tx, selectBuilder, preRegistrationStruct)
}

func (d *CommonDB) UpdatePreRegistration(tx *sql.Tx, preRegistration *models.PreRegistration) error {
	if preRegistration.Id == 0 {
		return errors.WithStack(errors.New("can't update preRegistration with id 0"))
	}

	originalUpdatedAt := preRegistration.UpdatedAt
	preRegistration.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	preRegistrationStruct := sqlbuilder.NewStruct(new(models.PreRegistration)).For(d.Flavor)
	updateBuilder := preRegistrationStruct.WithoutTag("pk").WithoutTag("dont-update").Update("pre_registrations", preRegistration)
	updateBuilder.Where(updateBuilder.Equal("id", preRegistration.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		preRegistration.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update preRegistration")
	}

	return nil
}

func (d *CommonDB) DeletePreRegistration(tx *sql.Tx, preRegistrationId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.PreRegistration)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("pre_registrations")
	deleteBuilder.Where(deleteBuilder.Equal("id", preRegistrationId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete preRegistration")
	}

	return nil
}

func (d *CommonDB) getPreRegistrationCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, preRegistrationStruct *sqlbuilder.Struct) (*models.PreRegistration, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var preRegistration models.PreRegistration
	if rows.Next() {
		addr := preRegistrationStruct.Addr(&preRegistration)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan preRegistration")
		}
		return &preRegistration, nil
	}

	return nil, nil
}
