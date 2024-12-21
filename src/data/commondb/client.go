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

func (d *CommonDB) GetClientsByIds(tx *sql.Tx, clientIds []int64) (clients []models.Client, err error) {
	if len(clientIds) == 0 {
		return []models.Client{}, nil
	}

	clientStruct := sqlbuilder.NewStruct(new(models.Client)).For(d.Flavor)
	selectBuilder := clientStruct.SelectFrom("clients")
	selectBuilder.Where(selectBuilder.In("id", sqlbuilder.Flatten(clientIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var client models.Client
		addr := clientStruct.Addr(&client)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan client")
		}
		clients = append(clients, client)
	}

	return
}

func (d *CommonDB) GetAllClients(tx *sql.Tx) (clients []models.Client, err error) {
	clientStruct := sqlbuilder.NewStruct(new(models.Client)).For(d.Flavor)
	selectBuilder := clientStruct.SelectFrom("clients")
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var client models.Client
		addr := clientStruct.Addr(&client)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan client")
		}
		clients = append(clients, client)
	}

	return
}

func (d *CommonDB) getClientCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, clientStruct *sqlbuilder.Struct) (*models.Client, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var client models.Client
	if rows.Next() {
		addr := clientStruct.Addr(&client)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan client")
		}
		return &client, nil
	}

	return nil, nil
}
