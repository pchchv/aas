package oauth

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pchchv/aas/pkg/src/database"
)

type TokenParser struct {
	database database.Database
}

func NewTokenParser(database database.Database) *TokenParser {
	return &TokenParser{
		database: database,
	}
}

func (tp *TokenParser) DecodeAndValidateTokenString(token string, pubKey *rsa.PublicKey, withExpirationCheck bool) (t *Jwt, err error) {
	if pubKey == nil {
		if pubKey, err = tp.getPublicKey(); err != nil {
			return nil, err
		}
	}

	t = &Jwt{
		TokenBase64: token,
	}

	if len(token) > 0 {
		claims, opts := jwt.MapClaims{}, []jwt.ParserOption{}
		if withExpirationCheck {
			opts = append(opts, jwt.WithExpirationRequired())
		} else {
			opts = append(opts, jwt.WithoutClaimsValidation())
		}

		if _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return pubKey, nil
		}, opts...); err != nil {
			return nil, err
		}
		t.Claims = claims
	}

	return
}

func (tp *TokenParser) DecodeAndValidateTokenResponse(tokenResponse *TokenResponse) (token *JwtInfo, err error) {
	pubKey, err := tp.getPublicKey()
	if err != nil {
		return nil, err
	}

	token = &JwtInfo{
		TokenResponse: *tokenResponse,
	}

	if len(tokenResponse.AccessToken) > 0 {
		if token.AccessToken, err = tp.DecodeAndValidateTokenString(tokenResponse.AccessToken, pubKey, true); err != nil {
			return nil, err
		}
	}

	if len(tokenResponse.IdToken) > 0 {
		if token.IdToken, err = tp.DecodeAndValidateTokenString(tokenResponse.IdToken, pubKey, true); err != nil {
			return nil, err
		}
	}

	if len(tokenResponse.RefreshToken) > 0 {
		if token.RefreshToken, err = tp.DecodeAndValidateTokenString(tokenResponse.RefreshToken, pubKey, false); err != nil {
			return nil, err
		}
	}

	return
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
