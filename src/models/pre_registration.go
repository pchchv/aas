package models

import "database/sql"

type PreRegistration struct {
	Id                        int64        `db:"id" fieldtag:"pk"`
	Email                     string       `db:"email"`
	CreatedAt                 sql.NullTime `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt                 sql.NullTime `db:"updated_at"`
	PasswordHash              string       `db:"password_hash"`
	VerificationCodeIssuedAt  sql.NullTime `db:"verification_code_issued_at"`
	VerificationCodeEncrypted []byte       `db:"verification_code_encrypted"`
}
