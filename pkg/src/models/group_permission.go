package models

import "database/sql"

type GroupPermission struct {
	Id           int64        `db:"id" fieldtag:"pk"`
	GroupId      int64        `db:"group_id"`
	CreatedAt    sql.NullTime `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
	PermissionId int64        `db:"permission_id"`
}
