package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateUserConsent(tx *sql.Tx, userConsent *models.UserConsent) error {
	if userConsent.ClientId == 0 {
		return errors.WithStack(errors.New("client id must be greater than 0"))
	}

	if userConsent.UserId == 0 {
		return errors.WithStack(errors.New("user id must be greater than 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := userConsent.CreatedAt
	originalUpdatedAt := userConsent.UpdatedAt
	userConsent.CreatedAt = sql.NullTime{Time: now, Valid: true}
	userConsent.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	userConsentStruct := sqlbuilder.NewStruct(new(models.UserConsent)).For(d.Flavor)
	insertBuilder := userConsentStruct.WithoutTag("pk").InsertInto("user_consents", userConsent)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		userConsent.CreatedAt = originalCreatedAt
		userConsent.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert userConsent")
	}

	id, err := result.LastInsertId()
	if err != nil {
		userConsent.CreatedAt = originalCreatedAt
		userConsent.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	userConsent.Id = id
	return nil
}

func (d *CommonDB) UpdateUserConsent(tx *sql.Tx, userConsent *models.UserConsent) error {
	if userConsent.Id == 0 {
		return errors.WithStack(errors.New("can't update userConsent with id 0"))
	}

	originalUpdatedAt := userConsent.UpdatedAt
	userConsent.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	userConsentStruct := sqlbuilder.NewStruct(new(models.UserConsent)).For(d.Flavor)
	updateBuilder := userConsentStruct.WithoutTag("pk").WithoutTag("dont-update").Update("user_consents", userConsent)
	updateBuilder.Where(updateBuilder.Equal("id", userConsent.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		userConsent.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update userConsent")
	}

	return nil
}

func (d *CommonDB) GetConsentsByUserId(tx *sql.Tx, userId int64) (userConsents []models.UserConsent, err error) {
	userConsentStruct := sqlbuilder.NewStruct(new(models.UserConsent)).For(d.Flavor)
	selectBuilder := userConsentStruct.SelectFrom("user_consents")
	selectBuilder.Where(selectBuilder.Equal("user_id", userId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userConsent models.UserConsent
		addr := userConsentStruct.Addr(&userConsent)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userConsent")
		}
		userConsents = append(userConsents, userConsent)
	}

	return
}

func (d *CommonDB) GetConsentByUserIdAndClientId(tx *sql.Tx, userId int64, clientId int64) (*models.UserConsent, error) {
	userConsentStruct := sqlbuilder.NewStruct(new(models.UserConsent)).For(d.Flavor)
	selectBuilder := userConsentStruct.SelectFrom("user_consents")
	selectBuilder.Where(selectBuilder.Equal("user_id", userId))
	selectBuilder.Where(selectBuilder.Equal("client_id", clientId))
	return d.getUserConsentCommon(tx, selectBuilder, userConsentStruct)
}

func (d *CommonDB) GetUserConsentById(tx *sql.Tx, userConsentId int64) (*models.UserConsent, error) {
	userConsentStruct := sqlbuilder.NewStruct(new(models.UserConsent)).For(d.Flavor)
	selectBuilder := userConsentStruct.SelectFrom("user_consents")
	selectBuilder.Where(selectBuilder.Equal("id", userConsentId))
	return d.getUserConsentCommon(tx, selectBuilder, userConsentStruct)
}

func (d *CommonDB) DeleteUserConsent(tx *sql.Tx, userConsentId int64) error {
	userConsentStruct := sqlbuilder.NewStruct(new(models.UserConsent)).For(d.Flavor)
	deleteBuilder := userConsentStruct.DeleteFrom("user_consents")
	deleteBuilder.Where(deleteBuilder.Equal("id", userConsentId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete userConsent")
	}

	return nil
}

func (d *CommonDB) DeleteAllUserConsent(tx *sql.Tx) error {
	userConsentStruct := sqlbuilder.NewStruct(new(models.UserConsent)).For(d.Flavor)
	deleteBuilder := userConsentStruct.DeleteFrom("user_consents")
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete userConsent")
	}

	return nil
}

func (d *CommonDB) getUserConsentCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, userConsentStruct *sqlbuilder.Struct) (*models.UserConsent, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var userConsent models.UserConsent
	if rows.Next() {
		addr := userConsentStruct.Addr(&userConsent)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userConsent")
		}
		return &userConsent, nil
	}

	return nil, nil
}
