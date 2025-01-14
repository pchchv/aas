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
