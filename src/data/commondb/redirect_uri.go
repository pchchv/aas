package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateRedirectURI(tx *sql.Tx, redirectURI *models.RedirectURI) error {
	if redirectURI.ClientId == 0 {
		return errors.WithStack(errors.New("client id must be greater than 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := redirectURI.CreatedAt
	redirectURI.CreatedAt = sql.NullTime{Time: now, Valid: true}
	redirectURIStruct := sqlbuilder.NewStruct(new(models.RedirectURI)).For(d.Flavor)
	insertBuilder := redirectURIStruct.WithoutTag("pk").InsertInto("redirect_uris", redirectURI)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		redirectURI.CreatedAt = originalCreatedAt
		return errors.Wrap(err, "unable to insert redirectURI")
	}

	id, err := result.LastInsertId()
	if err != nil {
		redirectURI.CreatedAt = originalCreatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	redirectURI.Id = id
	return nil
}

func (d *CommonDB) GetRedirectURIsByClientId(tx *sql.Tx, clientId int64) (redirectURIs []models.RedirectURI, err error) {
	redirectURIStruct := sqlbuilder.NewStruct(new(models.RedirectURI)).For(d.Flavor)
	selectBuilder := redirectURIStruct.SelectFrom("redirect_uris")
	selectBuilder.Where(selectBuilder.Equal("client_id", clientId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var redirectURI models.RedirectURI
		addr := redirectURIStruct.Addr(&redirectURI)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan redirectURI")
		}
		redirectURIs = append(redirectURIs, redirectURI)
	}

	return
}

func (d *CommonDB) DeleteRedirectURI(tx *sql.Tx, redirectURIId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.RedirectURI)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("redirect_uris")
	deleteBuilder.Where(deleteBuilder.Equal("id", redirectURIId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete redirectURI")
	}

	return nil
}
