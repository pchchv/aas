package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateWebOrigin(tx *sql.Tx, webOrigin *models.WebOrigin) error {
	if webOrigin.ClientId == 0 {
		return errors.WithStack(errors.New("client id must be greater than 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := webOrigin.CreatedAt
	webOrigin.CreatedAt = sql.NullTime{Time: now, Valid: true}
	webOriginStruct := sqlbuilder.NewStruct(new(models.WebOrigin)).For(d.Flavor)
	insertBuilder := webOriginStruct.WithoutTag("pk").InsertInto("web_origins", webOrigin)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		webOrigin.CreatedAt = originalCreatedAt
		return errors.Wrap(err, "unable to insert webOrigin")
	}

	id, err := result.LastInsertId()
	if err != nil {
		webOrigin.CreatedAt = originalCreatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	webOrigin.Id = id
	return nil
}

func (d *CommonDB) GetAllWebOrigins(tx *sql.Tx) (webOrigins []models.WebOrigin, err error) {
	webOriginStruct := sqlbuilder.NewStruct(new(models.WebOrigin)).For(d.Flavor)
	selectBuilder := webOriginStruct.SelectFrom("web_origins")
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var webOrigin models.WebOrigin
		addr := webOriginStruct.Addr(&webOrigin)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan webOrigin")
		}

		webOrigins = append(webOrigins, webOrigin)
	}

	return
}

func (d *CommonDB) GetWebOriginById(tx *sql.Tx, webOriginId int64) (*models.WebOrigin, error) {
	webOriginStruct := sqlbuilder.NewStruct(new(models.WebOrigin)).For(d.Flavor)
	selectBuilder := webOriginStruct.SelectFrom("web_origins")
	selectBuilder.Where(selectBuilder.Equal("id", webOriginId))
	return d.getWebOriginCommon(tx, selectBuilder, webOriginStruct)
}

func (d *CommonDB) GetWebOriginsByClientId(tx *sql.Tx, clientId int64) (webOrigins []models.WebOrigin, err error) {
	webOriginStruct := sqlbuilder.NewStruct(new(models.WebOrigin)).For(d.Flavor)
	selectBuilder := webOriginStruct.SelectFrom("web_origins")
	selectBuilder.Where(selectBuilder.Equal("client_id", clientId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var webOrigin models.WebOrigin
		addr := webOriginStruct.Addr(&webOrigin)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan webOrigin")
		}

		webOrigins = append(webOrigins, webOrigin)
	}

	return
}

func (d *CommonDB) DeleteWebOrigin(tx *sql.Tx, webOriginId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.WebOrigin)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("web_origins")
	deleteBuilder.Where(deleteBuilder.Equal("id", webOriginId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete webOrigin")
	}
	return nil
}

func (d *CommonDB) getWebOriginCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, webOriginStruct *sqlbuilder.Struct) (*models.WebOrigin, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var webOrigin models.WebOrigin
	if rows.Next() {
		addr := webOriginStruct.Addr(&webOrigin)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan webOrigin")
		}

		return &webOrigin, nil
	}

	return nil, nil
}
