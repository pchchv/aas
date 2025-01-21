package oauth

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt"
	db "github.com/pchchv/aas/src/data"
)

type TokenParser struct {
	database db.Database
}

func NewTokenParser(database db.Database) *TokenParser {
	return &TokenParser{
		database: database,
	}
}

func (tp *TokenParser) getPublicKey() (pubKey *rsa.PublicKey, err error) {
	keyPair, err := tp.database.GetCurrentSigningKey(nil)
	if err != nil {
		return nil, err
	}

	pubKey, err = jwt.ParseRSAPublicKeyFromPEM(keyPair.PublicKeyPEM)
	if err != nil {
		return nil, err
	}

	return
}
