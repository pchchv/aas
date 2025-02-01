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

type authHelper interface {
	RedirToAuthorize(w http.ResponseWriter, r *http.Request, clientIdentifier string, scope string, redirectBack string) error
	IsAuthorizedToAccessResource(jwtInfo oauth.JwtInfo, scopesAnyOf []string) bool
	IsAuthenticated(jwtInfo oauth.JwtInfo) bool
}
