package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateUserSessionClient(tx *sql.Tx, userSessionClient *models.UserSessionClient) error {
	now := time.Now().UTC()
	originalCreatedAt := userSessionClient.CreatedAt
	originalUpdatedAt := userSessionClient.UpdatedAt
	userSessionClient.CreatedAt = sql.NullTime{Time: now, Valid: true}
	userSessionClient.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	userSessionClientStruct := sqlbuilder.NewStruct(new(models.UserSessionClient)).For(d.Flavor)
	insertBuilder := userSessionClientStruct.WithoutTag("pk").InsertInto("user_session_clients", userSessionClient)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		userSessionClient.CreatedAt = originalCreatedAt
		userSessionClient.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert userSessionClient")
	}

	id, err := result.LastInsertId()
	if err != nil {
		userSessionClient.CreatedAt = originalCreatedAt
		userSessionClient.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	userSessionClient.Id = id
	return nil
}

func (d *CommonDB) GetUserSessionsClientByIds(tx *sql.Tx, userSessionClientIds []int64) (userSessionClients []models.UserSessionClient, err error) {
	if len(userSessionClientIds) == 0 {
		return nil, nil
	}

	userSessionClientStruct := sqlbuilder.NewStruct(new(models.UserSessionClient)).For(d.Flavor)
	selectBuilder := userSessionClientStruct.SelectFrom("user_session_clients")
	selectBuilder.Where(selectBuilder.In("id", sqlbuilder.Flatten(userSessionClientIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userSessionClient models.UserSessionClient
		addr := userSessionClientStruct.Addr(&userSessionClient)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userSessionClient")
		}
		userSessionClients = append(userSessionClients, userSessionClient)
	}

	return
}

func (d *CommonDB) GetUserSessionClientsByUserSessionIds(tx *sql.Tx, userSessionIds []int64) (userSessionClients []models.UserSessionClient, err error) {
	if len(userSessionIds) == 0 {
		return nil, nil
	}

	userSessionClientStruct := sqlbuilder.NewStruct(new(models.UserSessionClient)).For(d.Flavor)
	selectBuilder := userSessionClientStruct.SelectFrom("user_session_clients")
	selectBuilder.Where(selectBuilder.In("user_session_id", sqlbuilder.Flatten(userSessionIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userSessionClient models.UserSessionClient
		addr := userSessionClientStruct.Addr(&userSessionClient)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userSessionClient")
		}
		userSessionClients = append(userSessionClients, userSessionClient)
	}

	return
}
