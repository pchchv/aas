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

func (d *CommonDB) GetUserSessionClientById(tx *sql.Tx, userSessionClientId int64) (*models.UserSessionClient, error) {
	userSessionClientStruct := sqlbuilder.NewStruct(new(models.UserSessionClient)).For(d.Flavor)
	selectBuilder := userSessionClientStruct.SelectFrom("user_session_clients")
	selectBuilder.Where(selectBuilder.Equal("id", userSessionClientId))
	return d.getUserSessionClientCommon(tx, selectBuilder, userSessionClientStruct)
}

func (d *CommonDB) GetUserSessionClientsByUserSessionId(tx *sql.Tx, userSessionId int64) (userSessionClients []models.UserSessionClient, err error) {
	userSessionClientStruct := sqlbuilder.NewStruct(new(models.UserSessionClient)).For(d.Flavor)
	selectBuilder := userSessionClientStruct.SelectFrom("user_session_clients")
	selectBuilder.Where(selectBuilder.Equal("user_session_id", userSessionId))
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

func (d *CommonDB) UpdateUserSessionClient(tx *sql.Tx, userSessionClient *models.UserSessionClient) error {
	if userSessionClient.Id == 0 {
		return errors.WithStack(errors.New("can't update userSessionClient with id 0"))
	}

	originalUpdatedAt := userSessionClient.UpdatedAt
	userSessionClient.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	userSessionClientStruct := sqlbuilder.NewStruct(new(models.UserSessionClient)).For(d.Flavor)
	updateBuilder := userSessionClientStruct.WithoutTag("pk").WithoutTag("dont-update").Update("user_session_clients", userSessionClient)
	updateBuilder.Where(updateBuilder.Equal("id", userSessionClient.Id))

	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		userSessionClient.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update userSessionClient")
	}

	return nil
}

func (d *CommonDB) DeleteUserSessionClient(tx *sql.Tx, userSessionClientId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.UserSessionClient)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("user_session_clients")
	deleteBuilder.Where(deleteBuilder.Equal("id", userSessionClientId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete userSessionClient")
	}

	return nil
}

func (d *CommonDB) UserSessionClientsLoadClients(tx *sql.Tx, userSessionClients []models.UserSessionClient) error {
	if userSessionClients == nil {
		return nil
	}

	clientIds := make([]int64, 0)
	for _, userSessionClient := range userSessionClients {
		clientIds = append(clientIds, userSessionClient.ClientId)
	}

	clients, err := d.GetClientsByIds(tx, clientIds)
	if err != nil {
		return errors.Wrap(err, "unable to get clients by ids")
	}

	clientsMap := make(map[int64]models.Client)
	for _, client := range clients {
		clientsMap[client.Id] = client
	}

	for i, userSessionClient := range userSessionClients {
		if client, ok := clientsMap[userSessionClient.ClientId]; !ok {
			return errors.Errorf("client with id %d not found", userSessionClient.ClientId)
		} else {
			userSessionClients[i].Client = client
		}
	}

	return nil
}

func (d *CommonDB) getUserSessionClientCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, userSessionClientStruct *sqlbuilder.Struct) (*models.UserSessionClient, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var userSessionClient models.UserSessionClient
	if rows.Next() {
		addr := userSessionClientStruct.Addr(&userSessionClient)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userSessionClient")
		}

		return &userSessionClient, nil
	}

	return nil, nil
}
