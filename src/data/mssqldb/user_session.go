package mssqldb

import (
	"database/sql"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *MsSQLDB) CreateUserSession(tx *sql.Tx, userSession *models.UserSession) error {
	if userSession.UserId == 0 {
		return errors.WithStack(errors.New("user id must be greater than 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := userSession.CreatedAt
	originalUpdatedAt := userSession.UpdatedAt
	userSession.CreatedAt = sql.NullTime{Time: now, Valid: true}
	userSession.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	userSessionStruct := sqlbuilder.NewStruct(new(models.UserSession)).For(sqlbuilder.SQLServer)
	insertBuilder := userSessionStruct.WithoutTag("pk").InsertInto("user_sessions", userSession)
	sql, args := insertBuilder.Build()
	parts := strings.SplitN(sql, "VALUES", 2)
	if len(parts) != 2 {
		return errors.New("unexpected SQL format from sqlbuilder")
	}

	sql = parts[0] + "OUTPUT INSERTED.id VALUES" + parts[1]
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		userSession.CreatedAt = originalCreatedAt
		userSession.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert userSession")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&userSession.Id); err != nil {
			userSession.CreatedAt = originalCreatedAt
			userSession.UpdatedAt = originalUpdatedAt
			return errors.Wrap(err, "unable to scan userSession id")
		}
	}

	return nil
}

func (d *MsSQLDB) UpdateUserSession(tx *sql.Tx, userSession *models.UserSession) error {
	return d.CommonDB.UpdateUserSession(tx, userSession)
}

func (d *MsSQLDB) GetUserSessionById(tx *sql.Tx, userSessionId int64) (*models.UserSession, error) {
	return d.CommonDB.GetUserSessionById(tx, userSessionId)
}

func (d *MsSQLDB) GetUserSessionBySessionIdentifier(tx *sql.Tx, sessionIdentifier string) (*models.UserSession, error) {
	return d.CommonDB.GetUserSessionBySessionIdentifier(tx, sessionIdentifier)
}

func (d *MsSQLDB) GetUserSessionsByClientIdPaginated(tx *sql.Tx, clientId int64, page int, pageSize int) ([]models.UserSession, int, error) {
	return d.CommonDB.GetUserSessionsByClientIdPaginated(tx, clientId, page, pageSize)
}

func (d *MsSQLDB) GetUserSessionsByUserId(tx *sql.Tx, userId int64) ([]models.UserSession, error) {
	return d.CommonDB.GetUserSessionsByUserId(tx, userId)
}

func (d *MsSQLDB) DeleteUserSession(tx *sql.Tx, userSessionId int64) error {
	return d.CommonDB.DeleteUserSession(tx, userSessionId)
}

func (d *MsSQLDB) DeleteIdleSessions(tx *sql.Tx, idleTimeout time.Duration) error {
	return d.CommonDB.DeleteIdleSessions(tx, idleTimeout)
}

func (d *MsSQLDB) DeleteExpiredSessions(tx *sql.Tx, maxLifetime time.Duration) error {
	return d.CommonDB.DeleteExpiredSessions(tx, maxLifetime)
}

func (d *MsSQLDB) UserSessionsLoadUsers(tx *sql.Tx, userSessions []models.UserSession) error {
	return d.CommonDB.UserSessionsLoadUsers(tx, userSessions)
}

func (d *MsSQLDB) UserSessionsLoadClients(tx *sql.Tx, userSessions []models.UserSession) error {
	return d.CommonDB.UserSessionsLoadClients(tx, userSessions)
}

func (d *MsSQLDB) UserSessionLoadClients(tx *sql.Tx, userSession *models.UserSession) error {
	return d.CommonDB.UserSessionLoadClients(tx, userSession)
}

func (d *MsSQLDB) UserSessionLoadUser(tx *sql.Tx, userSession *models.UserSession) error {
	return d.CommonDB.UserSessionLoadUser(tx, userSession)
}
