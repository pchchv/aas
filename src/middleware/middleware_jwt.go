package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/src/config"
	"github.com/pchchv/aas/src/constants"
	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/encryption"
	"github.com/pchchv/aas/src/models"
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

func (m *MiddlewareJwt) refreshToken(w http.ResponseWriter, r *http.Request, tokenResponse *oauth.TokenResponse) (bool, error) {
	if tokenResponse.RefreshToken == "" {
		return false, nil
	}

	client, err := m.database.GetClientByClientIdentifier(nil, constants.AdminConsoleClientIdentifier)
	if err != nil {
		return false, errors.New("unable to get client: " + err.Error())
	} else if client == nil {
		return false, errors.New("client is nil in refreshToken (middleware_jwt)")
	}

	settings := r.Context().Value(constants.ContextKeySettings).(*models.Settings)
	clientSecretDecrypted, err := encryption.DecryptText(client.ClientSecretEncrypted, settings.AESEncryptionKey)
	if err != nil {
		return false, errors.New("unable to decrypt client secret: " + err.Error())
	}

	// Prepare the refresh token request
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", tokenResponse.RefreshToken)
	data.Set("client_id", constants.AdminConsoleClientIdentifier)
	data.Set("client_secret", clientSecretDecrypted)

	// Create the HTTP request
	req, err := http.NewRequest("POST", config.GetAuthServer().BaseURL+"/auth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return false, errors.New("error creating refresh token request: " + err.Error())
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	if m.httpClient == nil {
		slog.Error("http client is nil in refreshToken (middleware_jwt)")
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return false, errors.New("error sending refresh token request: " + err.Error())
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, errors.New("error reading refresh token response: " + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return false, errors.New("error response from server: " + string(body))
	}

	// Parse the new token response
	var newTokenResponse oauth.TokenResponse
	if err = json.Unmarshal(body, &newTokenResponse); err != nil {
		return false, errors.New("error parsing refresh token response: " + err.Error())
	}

	sess, err := m.sessionStore.Get(r, constants.SessionName)
	if err != nil {
		return false, errors.New("unable to get session: " + err.Error())
	}

	// Update the session with the new token response
	sess.Values[constants.SessionKeyJwt] = newTokenResponse
	if err = m.sessionStore.Save(r, w, sess); err != nil {
		return false, errors.New("unable to save the session: " + err.Error())
	}

	return true, nil
}
