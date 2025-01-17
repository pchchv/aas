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

func (d *CommonDB) GetRefreshTokenById(tx *sql.Tx, refreshTokenId int64) (*models.RefreshToken, error) {
	refreshTokenStruct := sqlbuilder.NewStruct(new(models.RefreshToken)).For(d.Flavor)
	selectBuilder := refreshTokenStruct.SelectFrom("refresh_tokens")
	selectBuilder.Where(selectBuilder.Equal("id", refreshTokenId))
	return d.getRefreshTokenCommon(tx, selectBuilder, refreshTokenStruct)
}

func (d *CommonDB) GetRefreshTokenByJti(tx *sql.Tx, jti string) (*models.RefreshToken, error) {
	refreshTokenStruct := sqlbuilder.NewStruct(new(models.RefreshToken)).For(d.Flavor)
	selectBuilder := refreshTokenStruct.SelectFrom("refresh_tokens")
	selectBuilder.Where(selectBuilder.Equal("refresh_token_jti", jti))
	return d.getRefreshTokenCommon(tx, selectBuilder, refreshTokenStruct)
}

func (d *CommonDB) UpdateRefreshToken(tx *sql.Tx, refreshToken *models.RefreshToken) error {
	if refreshToken.Id == 0 {
		return errors.WithStack(errors.New("can't update refreshToken with id 0"))
	}

	originalUpdatedAt := refreshToken.UpdatedAt
	refreshToken.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	refreshTokenStruct := sqlbuilder.NewStruct(new(models.RefreshToken)).For(d.Flavor)
	updateBuilder := refreshTokenStruct.WithoutTag("pk").WithoutTag("dont-update").Update("refresh_tokens", refreshToken)
	updateBuilder.Where(updateBuilder.Equal("id", refreshToken.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		refreshToken.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update refreshToken")
	}

	return nil
}

func (d *CommonDB) DeleteRefreshToken(tx *sql.Tx, refreshTokenId int64) error {
	userConsentStruct := sqlbuilder.NewStruct(new(models.RefreshToken)).For(d.Flavor)
	deleteBuilder := userConsentStruct.DeleteFrom("refresh_tokens")
	deleteBuilder.Where(deleteBuilder.Equal("id", refreshTokenId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete refreshToken")
	}

	return nil
}

func (d *CommonDB) DeleteExpiredOrRevokedRefreshTokens(tx *sql.Tx) error {
	deleteBuilder := d.Flavor.NewDeleteBuilder()
	deleteBuilder.DeleteFrom("refresh_tokens")
	now := time.Now().UTC()
	deleteBuilder.Where(
		deleteBuilder.Or(
			deleteBuilder.LessThan("expires_at", now),
			deleteBuilder.LessThan("max_lifetime", now),
			deleteBuilder.Equal("revoked", true),
		),
	)

	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete expired/revoked refresh tokens")
	}

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
