package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
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
