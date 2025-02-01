package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/src/constants"
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

// JwtAuthorizationHeaderToContext is a middleware that extracts the JWT token from the Authorization header and stores it in the context.
func (m *MiddlewareJwt) JwtAuthorizationHeaderToContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			const BEARER_SCHEMA = "Bearer "
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, BEARER_SCHEMA) && len(authHeader) >= len(BEARER_SCHEMA) {
				tokenStr := authHeader[len(BEARER_SCHEMA):]
				token, err := m.tokenParser.DecodeAndValidateTokenString(tokenStr, nil, true)
				if err == nil {
					ctx = context.WithValue(ctx, constants.ContextKeyBearerToken, *token)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
