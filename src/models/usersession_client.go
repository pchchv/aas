package models

import (
	"database/sql"
	"time"
)

type UserSessionClient struct {
	Id            int64        `db:"id" fieldtag:"pk"`
	Client        Client       `db:"-"`
	Started       time.Time    `db:"started"`
	ClientId      int64        `db:"client_id"`
	CreatedAt     sql.NullTime `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt     sql.NullTime `db:"updated_at"`
	LastAccessed  time.Time    `db:"last_accessed"`
	UserSessionId int64        `db:"user_session_id"`
}
