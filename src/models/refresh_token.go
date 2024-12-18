package models

import "database/sql"

type RefreshToken struct {
	Id                      int64        `db:"id" fieldtag:"pk"`
	Code                    Code         `db:"-"`
	Scope                   string       `db:"scope"`
	CodeId                  int64        `db:"code_id"`
	IssuedAt                sql.NullTime `db:"issued_at"`
	ExpiresAt               sql.NullTime `db:"expires_at"`
	CreatedAt               sql.NullTime `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt               sql.NullTime `db:"updated_at"`
	RefreshTokenJti         string       `db:"refresh_token_jti"`
	SessionIdentifier       string       `db:"session_identifier"`
	RefreshTokenType        string       `db:"refresh_token_type"`
	MaxLifetime             sql.NullTime `db:"max_lifetime"`
	Revoked                 bool         `db:"revoked"`
	FirstRefreshTokenJti    string       `db:"first_refresh_token_jti"`
	PreviousRefreshTokenJti string       `db:"previous_refresh_token_jti"`
}
