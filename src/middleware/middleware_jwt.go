package middleware

import (
	"crypto/rsa"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/oauth"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MiddlewareJwt struct {
	sessionStore sessions.Store
	tokenParser  tokenParser
	database     database.Database
	authHelper   authHelper
	httpClient   HTTPClient
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

func NewMiddlewareJwt(sessionStore sessions.Store, tokenParser tokenParser, database database.Database, authHelper authHelper, httpClient HTTPClient) *MiddlewareJwt {
	return &MiddlewareJwt{
		sessionStore: sessionStore,
		tokenParser:  tokenParser,
		database:     database,
		authHelper:   authHelper,
		httpClient:   httpClient,
	}
}
