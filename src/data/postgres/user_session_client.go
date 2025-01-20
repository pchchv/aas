package postgresdb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *PostgresDB) CreateUserSessionClient(tx *sql.Tx, userSessionClient *models.UserSessionClient) error {
	now := time.Now().UTC()
	originalCreatedAt := userSessionClient.CreatedAt
	originalUpdatedAt := userSessionClient.UpdatedAt
	userSessionClient.CreatedAt = sql.NullTime{Time: now, Valid: true}
	userSessionClient.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	userSessionClientStruct := sqlbuilder.NewStruct(new(models.UserSessionClient)).For(sqlbuilder.PostgreSQL)
	insertBuilder := userSessionClientStruct.WithoutTag("pk").InsertInto("user_session_clients", userSessionClient)
	sql, args := insertBuilder.Build()
	sql += " RETURNING id"
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		userSessionClient.CreatedAt = originalCreatedAt
		userSessionClient.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert userSessionClient")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&userSessionClient.Id); err != nil {
			userSessionClient.CreatedAt = originalCreatedAt
			userSessionClient.UpdatedAt = originalUpdatedAt
			return errors.Wrap(err, "unable to scan userSessionClient id")
		}
	}

	return nil
}

func (d *PostgresDB) UpdateUserSessionClient(tx *sql.Tx, userSessionClient *models.UserSessionClient) error {
	return d.CommonDB.UpdateUserSessionClient(tx, userSessionClient)
}

func (d *PostgresDB) GetUserSessionClientsByUserSessionIds(tx *sql.Tx, userSessionIds []int64) ([]models.UserSessionClient, error) {
	return d.CommonDB.GetUserSessionClientsByUserSessionIds(tx, userSessionIds)
}

func (d *PostgresDB) GetUserSessionClientsByUserSessionId(tx *sql.Tx, userSessionId int64) ([]models.UserSessionClient, error) {
	return d.CommonDB.GetUserSessionClientsByUserSessionId(tx, userSessionId)
}

func (d *PostgresDB) GetUserSessionsClientByIds(tx *sql.Tx, userSessionClientIds []int64) ([]models.UserSessionClient, error) {
	return d.CommonDB.GetUserSessionsClientByIds(tx, userSessionClientIds)
}

func (d *PostgresDB) GetUserSessionClientById(tx *sql.Tx, userSessionClientId int64) (*models.UserSessionClient, error) {
	return d.CommonDB.GetUserSessionClientById(tx, userSessionClientId)
}

func (d *PostgresDB) DeleteUserSessionClient(tx *sql.Tx, userSessionClientId int64) error {
	return d.CommonDB.DeleteUserSessionClient(tx, userSessionClientId)
}

func (d *PostgresDB) UserSessionClientsLoadClients(tx *sql.Tx, userSessionClients []models.UserSessionClient) error {
	return d.CommonDB.UserSessionClientsLoadClients(tx, userSessionClients)
}
