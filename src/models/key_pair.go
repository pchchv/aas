package models

import "database/sql"

type KeyPair struct {
	Id                int64        `db:"id" fieldtag:"pk"`
	Type              string       `db:"type" fieldopt:"withquote"`
	State             string       `db:"state"`
	CreatedAt         sql.NullTime `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt         sql.NullTime `db:"updated_at"`
	Algorithm         string       `db:"algorithm" fieldopt:"withquote"`
	KeyIdentifier     string       `db:"key_identifier"`
	PrivateKeyPEM     []byte       `db:"private_key_pem"`
	PublicKeyPEM      []byte       `db:"public_key_pem"`
	PublicKeyJWK      []byte       `db:"public_key_jwk"`
	PublicKeyASN1_DER []byte       `db:"public_key_asn1_der"`
}
