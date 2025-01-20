package mysqldb

import (
	"database/sql"
	"time"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateUserSession(tx *sql.Tx, userSession *models.UserSession) error {
	return d.CommonDB.CreateUserSession(tx, userSession)
}

func (d *MySQLDB) UpdateUserSession(tx *sql.Tx, userSession *models.UserSession) error {
	return d.CommonDB.UpdateUserSession(tx, userSession)
}

func (d *MySQLDB) GetUserSessionById(tx *sql.Tx, userSessionId int64) (*models.UserSession, error) {
	return d.CommonDB.GetUserSessionById(tx, userSessionId)
}

func (d *MySQLDB) GetUserSessionBySessionIdentifier(tx *sql.Tx, sessionIdentifier string) (*models.UserSession, error) {
	return d.CommonDB.GetUserSessionBySessionIdentifier(tx, sessionIdentifier)
}

func (d *MySQLDB) GetUserSessionsByClientIdPaginated(tx *sql.Tx, clientId int64, page int, pageSize int) ([]models.UserSession, int, error) {
	return d.CommonDB.GetUserSessionsByClientIdPaginated(tx, clientId, page, pageSize)
}

func (d *MySQLDB) GetUserSessionsByUserId(tx *sql.Tx, userId int64) ([]models.UserSession, error) {
	return d.CommonDB.GetUserSessionsByUserId(tx, userId)
}

func (d *MySQLDB) DeleteUserSession(tx *sql.Tx, userSessionId int64) error {
	return d.CommonDB.DeleteUserSession(tx, userSessionId)
}

func (d *MySQLDB) DeleteIdleSessions(tx *sql.Tx, idleTimeout time.Duration) error {
	return d.CommonDB.DeleteIdleSessions(tx, idleTimeout)
}

func (d *MySQLDB) DeleteExpiredSessions(tx *sql.Tx, maxLifetime time.Duration) error {
	return d.CommonDB.DeleteExpiredSessions(tx, maxLifetime)
}

func (d *MySQLDB) UserSessionsLoadUsers(tx *sql.Tx, userSessions []models.UserSession) error {
	return d.CommonDB.UserSessionsLoadUsers(tx, userSessions)
}

func (d *MySQLDB) UserSessionsLoadClients(tx *sql.Tx, userSessions []models.UserSession) error {
	return d.CommonDB.UserSessionsLoadClients(tx, userSessions)
}

func (d *MySQLDB) UserSessionLoadClients(tx *sql.Tx, userSession *models.UserSession) error {
	return d.CommonDB.UserSessionLoadClients(tx, userSession)
}

func (d *MySQLDB) UserSessionLoadUser(tx *sql.Tx, userSession *models.UserSession) error {
	return d.CommonDB.UserSessionLoadUser(tx, userSession)
}
