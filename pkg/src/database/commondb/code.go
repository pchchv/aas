package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
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

func (d *CommonDB) GetCodeByCodeHash(tx *sql.Tx, codeHash string, used bool) (*models.Code, error) {
	codeStruct := sqlbuilder.NewStruct(new(models.Code)).For(d.Flavor)
	selectBuilder := codeStruct.SelectFrom("codes")
	selectBuilder.Where(selectBuilder.Equal("code_hash", codeHash))
	selectBuilder.Where(selectBuilder.Equal("used", used))
	return d.getCodeCommon(tx, selectBuilder, codeStruct)
}

func (d *CommonDB) GetCodeById(tx *sql.Tx, codeId int64) (*models.Code, error) {
	codeStruct := sqlbuilder.NewStruct(new(models.Code)).For(d.Flavor)
	selectBuilder := codeStruct.SelectFrom("codes")
	selectBuilder.Where(selectBuilder.Equal("id", codeId))
	return d.getCodeCommon(tx, selectBuilder, codeStruct)
}

func (d *CommonDB) UpdateCode(tx *sql.Tx, code *models.Code) error {
	if code.Id == 0 {
		return errors.WithStack(errors.New("can't update code with id 0"))
	}

	originalUpdatedAt := code.UpdatedAt
	code.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	codeStruct := sqlbuilder.NewStruct(new(models.Code)).For(d.Flavor)
	updateBuilder := codeStruct.WithoutTag("pk").WithoutTag("dont-update").Update("codes", code)
	updateBuilder.Where(updateBuilder.Equal("id", code.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		code.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update code")
	}

	return nil
}

func (d *CommonDB) getCodeCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, codeStruct *sqlbuilder.Struct) (*models.Code, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var code models.Code
	if rows.Next() {
		addr := codeStruct.Addr(&code)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan code")
		}
		return &code, nil
	}

	return nil, nil
}

func (d *CommonDB) DeleteUsedCodesWithoutRefreshTokens(tx *sql.Tx) error {
	deleteBuilder := d.Flavor.NewDeleteBuilder()
	deleteBuilder.DeleteFrom("codes")
	deleteBuilder.Where(
		deleteBuilder.And(
			deleteBuilder.Equal("used", true),
			deleteBuilder.NotIn("id",
				d.Flavor.NewSelectBuilder().Select("code_id").From("refresh_tokens"),
			),
		),
	)

	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete used codes without refresh tokens")
	}

	return nil
}

func (d *CommonDB) CodeLoadClient(tx *sql.Tx, code *models.Code) error {
	if code != nil {
		if client, err := d.GetClientById(tx, code.ClientId); err != nil {
			return errors.Wrap(err, "unable to load client")
		} else if client != nil {
			code.Client = *client
		}
	}

	return nil
}

func (d *CommonDB) CodeLoadUser(tx *sql.Tx, code *models.Code) error {
	if code != nil {
		if user, err := d.GetUserById(tx, code.UserId); err != nil {
			return errors.Wrap(err, "unable to load user")
		} else if user != nil {
			code.User = *user
		}
	}

	return nil
}

func (d *CommonDB) DeleteCode(tx *sql.Tx, codeId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.Code)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("codes")
	deleteBuilder.Where(deleteBuilder.Equal("id", codeId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete code")
	}

	return nil
}
