package mssqldb

import (
	"database/sql"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

func (d *MsSQLDB) CreateUserConsent(tx *sql.Tx, userConsent *models.UserConsent) error {
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
	userConsentStruct := sqlbuilder.NewStruct(new(models.UserConsent)).For(sqlbuilder.SQLServer)
	insertBuilder := userConsentStruct.WithoutTag("pk").InsertInto("user_consents", userConsent)
	sql, args := insertBuilder.Build()
	parts := strings.SplitN(sql, "VALUES", 2)
	if len(parts) != 2 {
		return errors.New("unexpected SQL format from sqlbuilder")
	}

	sql = parts[0] + "OUTPUT INSERTED.id VALUES" + parts[1]
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		userConsent.CreatedAt = originalCreatedAt
		userConsent.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert userConsent")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&userConsent.Id); err != nil {
			userConsent.CreatedAt = originalCreatedAt
			userConsent.UpdatedAt = originalUpdatedAt
			return errors.Wrap(err, "unable to scan userConsent id")
		}
	}

	return nil
}

func (d *MsSQLDB) UpdateUserConsent(tx *sql.Tx, userConsent *models.UserConsent) error {
	return d.CommonDB.UpdateUserConsent(tx, userConsent)
}

func (d *MsSQLDB) GetUserConsentById(tx *sql.Tx, userConsentId int64) (*models.UserConsent, error) {
	return d.CommonDB.GetUserConsentById(tx, userConsentId)
}

func (d *MsSQLDB) GetConsentByUserIdAndClientId(tx *sql.Tx, userId int64, clientId int64) (*models.UserConsent, error) {
	return d.CommonDB.GetConsentByUserIdAndClientId(tx, userId, clientId)
}

func (d *MsSQLDB) GetConsentsByUserId(tx *sql.Tx, userId int64) ([]models.UserConsent, error) {
	return d.CommonDB.GetConsentsByUserId(tx, userId)
}

func (d *MsSQLDB) DeleteUserConsent(tx *sql.Tx, userConsentId int64) error {
	return d.CommonDB.DeleteUserConsent(tx, userConsentId)
}

func (d *MsSQLDB) DeleteAllUserConsent(tx *sql.Tx) error {
	return d.CommonDB.DeleteAllUserConsent(tx)
}

func (d *MsSQLDB) UserConsentsLoadClients(tx *sql.Tx, userConsents []models.UserConsent) error {
	return d.CommonDB.UserConsentsLoadClients(tx, userConsents)
}
