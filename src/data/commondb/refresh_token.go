package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateRefreshToken(tx *sql.Tx, refreshToken *models.RefreshToken) error {
	now := time.Now().UTC()
	originalCreatedAt := refreshToken.CreatedAt
	originalUpdatedAt := refreshToken.UpdatedAt
	refreshToken.CreatedAt = sql.NullTime{Time: now, Valid: true}
	refreshToken.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	refreshTokenStruct := sqlbuilder.NewStruct(new(models.RefreshToken)).For(d.Flavor)
	insertBuilder := refreshTokenStruct.WithoutTag("pk").InsertInto("refresh_tokens", refreshToken)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		refreshToken.CreatedAt = originalCreatedAt
		refreshToken.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert refreshToken")
	}

	id, err := result.LastInsertId()
	if err != nil {
		refreshToken.CreatedAt = originalCreatedAt
		refreshToken.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	refreshToken.Id = id
	return nil
}

func (d *CommonDB) getRefreshTokenCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, refreshTokenStruct *sqlbuilder.Struct) (*models.RefreshToken, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var refreshToken models.RefreshToken
	if rows.Next() {
		addr := refreshTokenStruct.Addr(&refreshToken)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan refreshToken")
		}
		return &refreshToken, nil
	}
	return nil, nil
}
