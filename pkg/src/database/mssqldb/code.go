package mssqldb

import (
	"database/sql"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

func (d *MsSQLDB) CreateCode(tx *sql.Tx, code *models.Code) error {
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
	codeStruct := sqlbuilder.NewStruct(new(models.Code)).For(sqlbuilder.SQLServer)
	insertBuilder := codeStruct.WithoutTag("pk").InsertInto("codes", code)
	sql, args := insertBuilder.Build()
	parts := strings.SplitN(sql, "VALUES", 2)
	if len(parts) != 2 {
		return errors.New("unexpected SQL format from sqlbuilder")
	}

	sql = parts[0] + "OUTPUT INSERTED.id VALUES" + parts[1]
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		code.CreatedAt = originalCreatedAt
		code.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert code")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&code.Id); err != nil {
			code.CreatedAt = originalCreatedAt
			code.UpdatedAt = originalUpdatedAt
			return errors.Wrap(err, "unable to scan code id")
		}
	}

	return nil
}

func (d *MsSQLDB) UpdateCode(tx *sql.Tx, code *models.Code) error {
	return d.CommonDB.UpdateCode(tx, code)
}

func (d *MsSQLDB) GetCodeById(tx *sql.Tx, codeId int64) (*models.Code, error) {
	return d.CommonDB.GetCodeById(tx, codeId)
}

func (d *MsSQLDB) GetCodeByCodeHash(tx *sql.Tx, codeHash string, used bool) (*models.Code, error) {
	return d.CommonDB.GetCodeByCodeHash(tx, codeHash, used)
}

func (d *MsSQLDB) DeleteCode(tx *sql.Tx, codeId int64) error {
	return d.CommonDB.DeleteCode(tx, codeId)
}

func (d *MsSQLDB) DeleteUsedCodesWithoutRefreshTokens(tx *sql.Tx) error {
	return d.CommonDB.DeleteUsedCodesWithoutRefreshTokens(tx)
}

func (d *MsSQLDB) CodeLoadClient(tx *sql.Tx, code *models.Code) error {
	return d.CommonDB.CodeLoadClient(tx, code)
}

func (d *MsSQLDB) CodeLoadUser(tx *sql.Tx, code *models.Code) error {
	return d.CommonDB.CodeLoadUser(tx, code)
}
