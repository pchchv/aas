package oauth

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt"
	"github.com/pchchv/aas/src/database"
)

type TokenParser struct {
	database database.Database
}

func NewTokenParser(database database.Database) *TokenParser {
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
