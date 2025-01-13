package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateCode(tx *sql.Tx, code *models.Code) error {
	if code.ClientId == 0 {
		return errors.WithStack(errors.New("client id must be greater than 0"))
	}

	if code.UserId == 0 {
		return errors.WithStack(errors.New("user id must be greater than 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := code.CreatedAt
	originalUpdatedAt := code.UpdatedAt
	code.CreatedAt = sql.NullTime{Time: now, Valid: true}
	code.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	codeStruct := sqlbuilder.NewStruct(new(models.Code)).For(d.Flavor)
	insertBuilder := codeStruct.WithoutTag("pk").InsertInto("codes", code)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		code.CreatedAt = originalCreatedAt
		code.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert code")
	}

	id, err := result.LastInsertId()
	if err != nil {
		code.CreatedAt = originalCreatedAt
		code.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	code.Id = id
	return nil
}
