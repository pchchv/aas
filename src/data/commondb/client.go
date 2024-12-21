package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateClient(tx *sql.Tx, client *models.Client) error {
	now := time.Now().UTC()
	originalCreatedAt := client.CreatedAt
	originalUpdatedAt := client.UpdatedAt
	client.CreatedAt = sql.NullTime{Time: now, Valid: true}
	client.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	clientStruct := sqlbuilder.NewStruct(new(models.Client)).For(d.Flavor)
	insertBuilder := clientStruct.WithoutTag("pk").InsertInto("clients", client)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		client.CreatedAt = originalCreatedAt
		client.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert client")
	}

	id, err := result.LastInsertId()
	if err != nil {
		client.CreatedAt = originalCreatedAt
		client.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	client.Id = id
	return nil
}

func (d *CommonDB) UpdateClient(tx *sql.Tx, client *models.Client) (err error) {
	if client.Id == 0 {
		return errors.WithStack(errors.New("can't update client with id 0"))
	}

	originalUpdatedAt := client.UpdatedAt
	client.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	clientStruct := sqlbuilder.NewStruct(new(models.Client)).For(d.Flavor)
	updateBuilder := clientStruct.WithoutTag("pk").WithoutTag("dont-update").Update("clients", client)
	updateBuilder.Where(updateBuilder.Equal("id", client.Id))
	sql, args := updateBuilder.Build()
	if _, err = d.ExecSql(tx, sql, args...); err != nil {
		client.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update client")
	}

	return nil
}
