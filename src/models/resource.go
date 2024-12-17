package models

import "database/sql"

type Resource struct {
	Id                 int64        `db:"id" fieldtag:"pk"`
	CreatedAt          sql.NullTime `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt          sql.NullTime `db:"updated_at"`
	ResourceIdentifier string       `db:"resource_identifier"`
	Description        string       `db:"description"`
}
