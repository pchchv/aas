package handlerhelpers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"

	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/src/config"
	"github.com/pchchv/aas/src/constants"
	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/hashutil"
	"github.com/pchchv/aas/src/oauth"
	"github.com/pchchv/aas/src/stringutil"
)

type AuthHelper struct {
	sessionStore sessions.Store
}

func NewAuthHelper(sessionStore sessions.Store) *AuthHelper {
	return &AuthHelper{
		sessionStore: sessionStore,
	}
}

func (s *AuthHelper) GetAuthContext(r *http.Request) (*oauth.AuthContext, error) {
	sess, err := s.sessionStore.Get(r, constants.SessionName)
	if err != nil {
		return nil, err
	}

	var authContext oauth.AuthContext
	if jsonData, ok := sess.Values[constants.SessionKeyAuthContext].(string); !ok {
		return nil, customerrors.ErrNoAuthContext
	} else if err = json.Unmarshal([]byte(jsonData), &authContext); err != nil {
		return nil, err
	}

	return &authContext, nil
}

func (s *AuthHelper) GetLoggedInSubject(r *http.Request) string {
	if r.Context().Value(constants.ContextKeyJwtInfo) != nil {
		if jwtInfo, ok := r.Context().Value(constants.ContextKeyJwtInfo).(oauth.JwtInfo); !ok {
			stackBytes := debug.Stack()
			slog.Error("unable to cast jwtInfo\n" + string(stackBytes))
			return ""
		} else if jwtInfo.IdToken != nil {
			return jwtInfo.IdToken.GetStringClaim("sub")
		}
	}

	return ""
}

func (s *AuthHelper) RedirToAuthorize(w http.ResponseWriter, r *http.Request, clientIdentifier string, scope string, redirectBack string) error {
	sess, err := s.sessionStore.Get(r, constants.SessionName)
	if err != nil {
		return err
	}

	redirectURI := config.Get().BaseURL + "/auth/callback"
	codeVerifier := stringutil.GenerateSecurityRandomString(120)
	codeChallenge := oauth.GeneratePKCECodeChallenge(codeVerifier)
	state := stringutil.GenerateSecurityRandomString(16)
	nonce := stringutil.GenerateSecurityRandomString(16)
	sess.Values[constants.SessionKeyState] = state
	sess.Values[constants.SessionKeyNonce] = nonce
	sess.Values[constants.SessionKeyCodeVerifier] = codeVerifier
	sess.Values[constants.SessionKeyRedirectURI] = redirectURI
	sess.Values[constants.SessionKeyRedirectBack] = redirectBack
	if err = s.sessionStore.Save(r, w, sess); err != nil {
		return err
	}

	values := url.Values{}
	values.Add("client_id", clientIdentifier)
	values.Add("redirect_uri", redirectURI)
	values.Add("response_mode", "form_post")
	values.Add("response_type", "code")
	values.Add("code_challenge_method", "S256")
	values.Add("code_challenge", codeChallenge)
	values.Add("state", state)
	if nonceHash, err := hashutil.HashString(nonce); err != nil {
		return err
	} else {
		values.Add("nonce", nonceHash)
		values.Add("scope", scope)
		values.Add("acr_values", "2") // pwd + optional otp (if enabled)
	}

	destUrl := config.GetAuthServer().BaseURL + "/auth/authorize?" + values.Encode()
	http.Redirect(w, r, destUrl, http.StatusFound)

	return nil
}

func (s *AuthHelper) ClearAuthContext(w http.ResponseWriter, r *http.Request) error {
	sess, err := s.sessionStore.Get(r, constants.SessionName)
	if err != nil {
		return err
	}

	delete(sess.Values, constants.SessionKeyAuthContext)

	return s.sessionStore.Save(r, w, sess)
}

func (s *AuthHelper) SaveAuthContext(w http.ResponseWriter, r *http.Request, authContext *oauth.AuthContext) error {
	sess, err := s.sessionStore.Get(r, constants.SessionName)
	if err != nil {
		return err
	}

	if jsonData, err := json.Marshal(authContext); err != nil {
		return err
	} else {
		sess.Values[constants.SessionKeyAuthContext] = string(jsonData)
	}

	return s.sessionStore.Save(r, w, sess)
}

func (s *AuthHelper) IsAuthorizedToAccessResource(jwtInfo oauth.JwtInfo, scopesAnyOf []string) bool {
	if jwtInfo.AccessToken != nil {
		for _, scope := range scopesAnyOf {
			if jwtInfo.AccessToken.HasScope(scope) {
				return true
			}
		}
	}
	return false
}

func (s *AuthHelper) IsAuthenticated(jwtInfo oauth.JwtInfo) bool {
	return jwtInfo.IdToken != nil
}
