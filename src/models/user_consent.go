package models

import "database/sql"

type UserConsent struct {
	Id        int64        `db:"id" fieldtag:"pk"`
	Scope     string       `db:"scope"`
	UserId    int64        `db:"user_id"`
	Client    Client       `db:"-"`
	ClientId  int64        `db:"client_id"`
	CreatedAt sql.NullTime `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt sql.NullTime `db:"updated_at"`
	GrantedAt sql.NullTime `db:"granted_at"`
}
