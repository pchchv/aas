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
	"sort"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/pkg/src/config"
	"github.com/pchchv/aas/pkg/src/constants"
	"github.com/pchchv/aas/pkg/src/database"
	"github.com/pchchv/aas/pkg/src/encryption"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pchchv/aas/pkg/src/oauth"
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

// JwtSessionHandler is a middleware that checks if the user has a valid JWT session.
// It will also refresh the token if needed.
func (m *MiddlewareJwt) JwtSessionHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			sess, err := m.sessionStore.Get(r, constants.SessionName)
			if err != nil {
				err = errors.New("unable to get the session: " + err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if sess.Values[constants.SessionKeyJwt] != nil {
				tokenResponse, ok := sess.Values[constants.SessionKeyJwt].(oauth.TokenResponse)
				if !ok {
					http.Error(w, "unable to cast the session value to TokenResponse", http.StatusInternalServerError)
					return
				}

				// Check if token needs refresh
				if _, err := m.tokenParser.DecodeAndValidateTokenString(tokenResponse.AccessToken, nil, true); err != nil {
					if refreshed, err := m.refreshToken(w, r, &tokenResponse); err != nil || !refreshed {
						// If refresh failed, clear the session and continue
						delete(sess.Values, constants.SessionKeyJwt)
						if err := m.sessionStore.Save(r, w, sess); err != nil {
							err = errors.New("unable to save the session: " + err.Error())
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
						next.ServeHTTP(w, r)
						return
					}
				}

				// Get the latest token response from the session
				tokenResponse = sess.Values[constants.SessionKeyJwt].(oauth.TokenResponse)
				if jwtInfo, err := m.tokenParser.DecodeAndValidateTokenResponse(&tokenResponse); err == nil {
					settings := r.Context().Value(constants.ContextKeySettings).(*models.Settings)
					// Check if any token has an invalid issuer
					hasInvalidIssuer := (jwtInfo.IdToken != nil && !jwtInfo.IdToken.IsIssuerValid(settings.Issuer)) ||
						(jwtInfo.AccessToken != nil && !jwtInfo.AccessToken.IsIssuerValid(settings.Issuer)) ||
						(jwtInfo.RefreshToken != nil && !jwtInfo.RefreshToken.IsIssuerValid(settings.Issuer))
					if hasInvalidIssuer {
						slog.Error("Invalid issuer in JWT token. Will clear the session and redirect to root")
						// Clear the session
						delete(sess.Values, constants.SessionKeyJwt)
						if err := m.sessionStore.Save(r, w, sess); err != nil {
							err = errors.New("unable to save the session: " + err.Error())
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}

						// Redirect to root
						http.Redirect(w, r, "/", http.StatusFound)
						return
					}

					ctx = context.WithValue(ctx, constants.ContextKeyJwtInfo, *jwtInfo)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequiresScope is a middleware that checks if the user has the required scope to access the resource.
func (m *MiddlewareJwt) RequiresScope(scopesAnyOf []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ok bool
			ctx := r.Context()
			var jwtInfo oauth.JwtInfo
			if r.Context().Value(constants.ContextKeyJwtInfo) != nil {
				if jwtInfo, ok = r.Context().Value(constants.ContextKeyJwtInfo).(oauth.JwtInfo); !ok {
					http.Error(w, "unable to cast the context value to JwtInfo in RequiresScope middleware", http.StatusInternalServerError)
					return
				}
			}

			if isAuthorized := m.authHelper.IsAuthorizedToAccessResource(jwtInfo, scopesAnyOf); !isAuthorized {
				if m.authHelper.IsAuthenticated(jwtInfo) {
					// User is authenticated but not authorized
					// Show the unauthorized page
					http.Redirect(w, r, "/unauthorized", http.StatusFound)
				} else {
					// User is not authenticated
					// Redirect to the authorize endpoint
					err := m.authHelper.RedirToAuthorize(w, r, constants.AdminConsoleClientIdentifier, m.buildScopeString(scopesAnyOf), config.Get().BaseURL+r.RequestURI)
					if err != nil {
						err = errors.New("unable to redirect to authorize in RequiresScope middleware: " + err.Error())
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
				return
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

func (m *MiddlewareJwt) buildScopeString(customScopes []string) string {
	scopeMap := make(map[string]bool)
	// Default required scopes
	defaultScopes := []string{
		"openid",
		"email",
		constants.AdminConsoleResourceIdentifier + ":" + constants.ManageAccountPermissionIdentifier,
		constants.AdminConsoleResourceIdentifier + ":" + constants.ManageAdminConsolePermissionIdentifier,
	}

	// Add default scopes first
	for _, scope := range defaultScopes {
		scopeMap[strings.ToLower(scope)] = true
	}

	// Add custom scopes
	for _, scope := range customScopes {
		if scope = strings.ToLower(strings.TrimSpace(scope)); scope != "" {
			scopeMap[scope] = true
		}
	}

	var allScopes []string
	for scope := range scopeMap {
		allScopes = append(allScopes, scope)
	}
	sort.Strings(allScopes)

	return strings.Join(allScopes, " ")
}
