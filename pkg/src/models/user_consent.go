package models

import (
	"database/sql"
	"slices"
	"strings"
)

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

func (uc *UserConsent) HasScope(scope string) bool {
	if len(uc.Scope) > 0 {
		return slices.Contains(strings.Split(uc.Scope, " "), scope)
	}
	return false
}
