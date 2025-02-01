package middleware

import (
	"crypto/rsa"
	"net/http"

	"github.com/pchchv/aas/src/oauth"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type tokenParser interface {
	DecodeAndValidateTokenResponse(tokenResponse *oauth.TokenResponse) (*oauth.JwtInfo, error)
	DecodeAndValidateTokenString(token string, pubKey *rsa.PublicKey, withExpirationCheck bool) (*oauth.Jwt, error)
}
