package helpers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/pkg/src/constants"
	"github.com/pchchv/aas/pkg/src/customerrors"
	"github.com/pchchv/aas/pkg/src/oauth"
	"github.com/stretchr/testify/assert"
)

func TestGetAuthContext(t *testing.T) {
	store := sessions.NewCookieStore([]byte("secret"))
	authHelper := NewAuthHelper(store)

	req, _ := http.NewRequest("GET", "/", nil)
	sess, _ := store.Get(req, constants.SessionName)
	authContext := &oauth.AuthContext{ClientId: "test_token"}
	jsonData, _ := json.Marshal(authContext)
	sess.Values[constants.SessionKeyAuthContext] = string(jsonData)
	sess.Save(req, nil)

	ctx, err := authHelper.GetAuthContext(req)
	assert.NoError(t, err)
	assert.Equal(t, authContext.ClientId, ctx.ClientId)
}

func TestGetAuthContext_NoAuthContext(t *testing.T) {
	store := sessions.NewCookieStore([]byte("secret"))
	authHelper := NewAuthHelper(store)

	req, _ := http.NewRequest("GET", "/", nil)
	sess, _ := store.Get(req, constants.SessionName)
	sess.Save(req, nil)

	ctx, err := authHelper.GetAuthContext(req)
	assert.Error(t, err)
	assert.Nil(t, ctx)
	assert.Equal(t, customerrors.ErrNoAuthContext, err)
}

func TestGetLoggedInSubject(t *testing.T) {
	store := sessions.NewCookieStore([]byte("secret"))
	authHelper := NewAuthHelper(store)

	req, _ := http.NewRequest("GET", "/", nil)
	jwtInfo := oauth.JwtInfo{IdToken: &oauth.Jwt{Claims: map[string]interface{}{"sub": "test_subject"}}}
	ctx := req.Context()
	ctx = context.WithValue(ctx, constants.ContextKeyJwtInfo, jwtInfo)
	req = req.WithContext(ctx)

	subject := authHelper.GetLoggedInSubject(req)
	assert.Equal(t, "test_subject", subject)
}

func TestRedirToAuthorize(t *testing.T) {
	store := sessions.NewCookieStore([]byte("secret"))
	authHelper := NewAuthHelper(store)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	err := authHelper.RedirToAuthorize(w, req, "client_id", "scope", "redirect_back")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, w.Code)
}

func TestClearAuthContext(t *testing.T) {
	store := sessions.NewCookieStore([]byte("secret"))
	authHelper := NewAuthHelper(store)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(req, constants.SessionName)
	sess.Values[constants.SessionKeyAuthContext] = "test"
	sess.Save(req, nil)

	err := authHelper.ClearAuthContext(w, req)
	assert.NoError(t, err)

	sess, _ = store.Get(req, constants.SessionName)
	_, ok := sess.Values[constants.SessionKeyAuthContext]
	assert.False(t, ok)
}

func TestSaveAuthContext(t *testing.T) {
	store := sessions.NewCookieStore([]byte("secret"))
	authHelper := NewAuthHelper(store)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	authContext := &oauth.AuthContext{ClientId: "test_token"}

	err := authHelper.SaveAuthContext(w, req, authContext)
	assert.NoError(t, err)

	sess, _ := store.Get(req, constants.SessionName)
	jsonData, _ := json.Marshal(authContext)
	assert.Equal(t, string(jsonData), sess.Values[constants.SessionKeyAuthContext])
}

func TestIsAuthorizedToAccessResource(t *testing.T) {
	store := sessions.NewCookieStore([]byte("secret"))
	authHelper := NewAuthHelper(store)

	jwtInfo := oauth.JwtInfo{IdToken: &oauth.Jwt{Claims: map[string]interface{}{"scope": []string{"scope1", "scope2"}}}}
	scopes := []string{"scope1"}

	isAuthorized := authHelper.IsAuthorizedToAccessResource(jwtInfo, scopes)
	assert.True(t, isAuthorized)
}

func TestIsAuthenticated(t *testing.T) {
	store := sessions.NewCookieStore([]byte("secret"))
	authHelper := NewAuthHelper(store)

	jwtInfo := oauth.JwtInfo{IdToken: &oauth.Jwt{}}

	isAuthenticated := authHelper.IsAuthenticated(jwtInfo)
	assert.True(t, isAuthenticated)
}
