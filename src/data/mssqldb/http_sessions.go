package mssqldb

import (
	"database/sql"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *MsSQLDB) CreateHttpSession(tx *sql.Tx, httpSession *models.HttpSession) error {
	now := time.Now().UTC()
	originalCreatedAt := httpSession.CreatedAt
	originalUpdatedAt := httpSession.UpdatedAt
	httpSession.CreatedAt = sql.NullTime{Time: now, Valid: true}
	httpSession.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	httpSessionStruct := sqlbuilder.NewStruct(new(models.HttpSession)).For(sqlbuilder.SQLServer)
	insertBuilder := httpSessionStruct.WithoutTag("pk").InsertInto("http_sessions", httpSession)
	sql, args := insertBuilder.Build()
	parts := strings.SplitN(sql, "VALUES", 2)
	if len(parts) != 2 {
		return errors.New("unexpected SQL format from sqlbuilder")
	}

	sql = parts[0] + "OUTPUT INSERTED.id VALUES" + parts[1]
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		httpSession.CreatedAt = originalCreatedAt
		httpSession.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert httpSession")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&httpSession.Id); err != nil {
			httpSession.CreatedAt = originalCreatedAt
			httpSession.UpdatedAt = originalUpdatedAt
			return errors.Wrap(err, "unable to scan httpSession id")
		}
	}

	return nil
}

func (d *MsSQLDB) UpdateHttpSession(tx *sql.Tx, httpSession *models.HttpSession) error {
	return d.CommonDB.UpdateHttpSession(tx, httpSession)
}

func (d *MsSQLDB) GetHttpSessionById(tx *sql.Tx, httpSessionId int64) (*models.HttpSession, error) {
	return d.CommonDB.GetHttpSessionById(tx, httpSessionId)
}

func (d *MsSQLDB) DeleteHttpSession(tx *sql.Tx, httpSessionId int64) error {
	return d.CommonDB.DeleteHttpSession(tx, httpSessionId)
}

func (d *MsSQLDB) DeleteHttpSessionExpired(tx *sql.Tx) error {
	return d.CommonDB.DeleteHttpSessionExpired(tx)
}
