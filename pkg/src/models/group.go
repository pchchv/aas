package models

import "database/sql"

type Group struct {
	Id                   int64            `db:"id" fieldtag:"pk"`
	CreatedAt            sql.NullTime     `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt            sql.NullTime     `db:"updated_at"`
	Attributes           []GroupAttribute `db:"-"`
	Description          string           `db:"description"`
	Permissions          []Permission     `db:"-"`
	GroupIdentifier      string           `db:"group_identifier"`
	IncludeInIdToken     bool             `db:"include_in_id_token"`
	IncludeInAccessToken bool             `db:"include_in_access_token"`
}
