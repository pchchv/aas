package handlerhelpers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/src/constants"
	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/oauth"
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
