package models

import "database/sql"

type RedirectURI struct {
	Id        int64        `db:"id" fieldtag:"pk"`
	URI       string       `db:"uri"`
	ClientId  int64        `db:"client_id"`
	CreatedAt sql.NullTime `db:"created_at" fieldtag:"dont-update"`
}
