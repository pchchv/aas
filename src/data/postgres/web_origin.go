package postgresdb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *PostgresDB) CreateWebOrigin(tx *sql.Tx, webOrigin *models.WebOrigin) error {
	if webOrigin.ClientId == 0 {
		return errors.WithStack(errors.New("client id must be greater than 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := webOrigin.CreatedAt
	webOrigin.CreatedAt = sql.NullTime{Time: now, Valid: true}
	webOriginStruct := sqlbuilder.NewStruct(new(models.WebOrigin)).For(sqlbuilder.PostgreSQL)
	insertBuilder := webOriginStruct.WithoutTag("pk").InsertInto("web_origins", webOrigin)
	sql, args := insertBuilder.Build()
	sql += " RETURNING id"
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		webOrigin.CreatedAt = originalCreatedAt
		return errors.Wrap(err, "unable to insert webOrigin")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&webOrigin.Id); err != nil {
			webOrigin.CreatedAt = originalCreatedAt
			return errors.Wrap(err, "unable to scan webOrigin id")
		}
	}

	return nil
}

func (d *PostgresDB) GetWebOriginById(tx *sql.Tx, webOriginId int64) (*models.WebOrigin, error) {
	return d.CommonDB.GetWebOriginById(tx, webOriginId)
}

func (d *PostgresDB) GetWebOriginsByClientId(tx *sql.Tx, clientId int64) ([]models.WebOrigin, error) {
	return d.CommonDB.GetWebOriginsByClientId(tx, clientId)
}

func (d *PostgresDB) GetAllWebOrigins(tx *sql.Tx) ([]models.WebOrigin, error) {
	return d.CommonDB.GetAllWebOrigins(tx)
}

func (d *PostgresDB) DeleteWebOrigin(tx *sql.Tx, webOriginId int64) error {
	return d.CommonDB.DeleteWebOrigin(tx, webOriginId)
}
