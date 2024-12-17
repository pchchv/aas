package models

import "database/sql"

type GroupAttribute struct {
	Id                   int64        `db:"id" fieldtag:"pk"`
	Key                  string       `db:"key" fieldopt:"withquote"`
	Value                string       `db:"value" fieldopt:"withquote"`
	GroupId              int64        `db:"group_id"`
	CreatedAt            sql.NullTime `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt            sql.NullTime `db:"updated_at"`
	IncludeInIdToken     bool         `db:"include_in_id_token"`
	IncludeInAccessToken bool         `db:"include_in_access_token"`
}
