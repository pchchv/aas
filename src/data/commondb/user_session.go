package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateUserSession(tx *sql.Tx, userSession *models.UserSession) error {
	if userSession.UserId == 0 {
		return errors.WithStack(errors.New("user id must be greater than 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := userSession.CreatedAt
	originalUpdatedAt := userSession.UpdatedAt
	userSession.CreatedAt = sql.NullTime{Time: now, Valid: true}
	userSession.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	userSessionStruct := sqlbuilder.NewStruct(new(models.UserSession)).For(d.Flavor)
	insertBuilder := userSessionStruct.WithoutTag("pk").InsertInto("user_sessions", userSession)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		userSession.CreatedAt = originalCreatedAt
		userSession.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert userSession")
	}

	if id, err := result.LastInsertId(); err != nil {
		userSession.CreatedAt = originalCreatedAt
		userSession.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	} else {
		userSession.Id = id
	}

	return nil
}

func (d *CommonDB) GetUserSessionsByClientIdPaginated(tx *sql.Tx, clientId int64, page int, pageSize int) (userSessions []models.UserSession, total int, err error) {
	if clientId <= 0 {
		return nil, 0, errors.WithStack(errors.New("client id must be greater than 0"))
	}

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	userSessionStruct := sqlbuilder.NewStruct(new(models.UserSession)).For(d.Flavor)
	selectBuilder := userSessionStruct.SelectFrom("user_sessions")
	selectBuilder.JoinWithOption(sqlbuilder.InnerJoin, "user_session_clients", "user_sessions.id = user_session_clients.user_session_id")
	selectBuilder.Where(selectBuilder.Equal("user_session_clients.client_id", clientId))
	selectBuilder.OrderBy("user_sessions.last_accessed").Desc()
	selectBuilder.Offset((page - 1) * pageSize)
	selectBuilder.Limit(pageSize)
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userSession models.UserSession
		addr := userSessionStruct.Addr(&userSession)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, 0, errors.Wrap(err, "unable to scan userSession")
		}
		userSessions = append(userSessions, userSession)
	}

	selectBuilder = d.Flavor.NewSelectBuilder()
	selectBuilder.Select("count(*)").From("user_sessions")
	selectBuilder.JoinWithOption(sqlbuilder.InnerJoin, "user_session_clients", "user_sessions.id = user_session_clients.user_session_id")
	selectBuilder.Where(selectBuilder.Equal("user_session_clients.client_id", clientId))

	sql, args = selectBuilder.Build()
	rows2, err := d.QuerySql(nil, sql, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to query database")
	}
	defer rows2.Close()

	if rows2.Next() {
		if err = rows2.Scan(&total); err != nil {
			return nil, 0, errors.Wrap(err, "unable to scan total")
		}
	}

	return
}

func (d *CommonDB) GetUserSessionsByUserId(tx *sql.Tx, userId int64) (userSessions []models.UserSession, err error) {
	userSessionStruct := sqlbuilder.NewStruct(new(models.UserSession)).For(d.Flavor)
	selectBuilder := userSessionStruct.SelectFrom("user_sessions")
	selectBuilder.Where(selectBuilder.Equal("user_id", userId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userSession models.UserSession
		addr := userSessionStruct.Addr(&userSession)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userSession")
		}
		userSessions = append(userSessions, userSession)
	}

	return
}

func (d *CommonDB) GetUserSessionById(tx *sql.Tx, userSessionId int64) (*models.UserSession, error) {
	userSessionStruct := sqlbuilder.NewStruct(new(models.UserSession)).For(d.Flavor)
	selectBuilder := userSessionStruct.SelectFrom("user_sessions")
	selectBuilder.Where(selectBuilder.Equal("id", userSessionId))
	return d.getUserSessionCommon(tx, selectBuilder, userSessionStruct)
}

func (d *CommonDB) GetUserSessionBySessionIdentifier(tx *sql.Tx, sessionIdentifier string) (*models.UserSession, error) {
	if sessionIdentifier == "" {
		return nil, nil
	}

	userSessionStruct := sqlbuilder.NewStruct(new(models.UserSession)).For(d.Flavor)
	selectBuilder := userSessionStruct.SelectFrom("user_sessions")
	selectBuilder.Where(selectBuilder.Equal("session_identifier", sessionIdentifier))
	return d.getUserSessionCommon(tx, selectBuilder, userSessionStruct)
}

func (d *CommonDB) getUserSessionCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, userSessionStruct *sqlbuilder.Struct) (*models.UserSession, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var userSession models.UserSession
	if rows.Next() {
		addr := userSessionStruct.Addr(&userSession)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userSession")
		}
		return &userSession, nil
	}

	return nil, nil
}
