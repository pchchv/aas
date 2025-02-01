package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/src/constants"
	dataMocks "github.com/pchchv/aas/src/database/mocks"
	"github.com/pchchv/aas/src/encryption"
	helpersMocks "github.com/pchchv/aas/src/helpers/mocks"
	"github.com/pchchv/aas/src/models"
	"github.com/pchchv/aas/src/oauth"
	OAuthMocks "github.com/pchchv/aas/src/oauth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestJwtAuthorizationHeaderToContext_ValidBearerToken(t *testing.T) {
	mockTokenParser := new(OAuthMocks.TokenParser)
	mockDatabase := new(dataMocks.Database)
	mockAuthHelper := new(helpersMocks.AuthHelper)
	middleware := NewMiddlewareJwt(nil, mockTokenParser, mockDatabase, mockAuthHelper, nil)
	expectedToken := &oauth.Jwt{
		TokenBase64: "validtoken",
		Claims: map[string]interface{}{
			"sub": "user",
		},
	}
	mockTokenParser.On("DecodeAndValidateTokenString", "validtoken", mock.Anything, true).
		Return(expectedToken, nil)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value(constants.ContextKeyBearerToken)
		assert.NotNil(t, token)
		assert.IsType(t, oauth.Jwt{}, token)
		assert.Equal(t, "validtoken", token.(oauth.Jwt).TokenBase64)
		assert.Equal(t, "user", token.(oauth.Jwt).Claims["sub"])
	})

	handler := middleware.JwtAuthorizationHeaderToContext()(nextHandler)
	handler.ServeHTTP(rr, req)

	mockTokenParser.AssertExpectations(t)
}

func TestJwtAuthorizationHeaderToContext_InvalidBearerToken(t *testing.T) {
	mockTokenParser := new(OAuthMocks.TokenParser)
	mockDatabase := new(dataMocks.Database)
	mockAuthHelper := new(helpersMocks.AuthHelper)
	middleware := NewMiddlewareJwt(nil, mockTokenParser, mockDatabase, mockAuthHelper, nil)
	mockTokenParser.On("DecodeAndValidateTokenString", "invalidtoken", mock.Anything, true).
		Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value(constants.ContextKeyBearerToken)
		assert.Nil(t, token)
	})

	handler := middleware.JwtAuthorizationHeaderToContext()(nextHandler)
	handler.ServeHTTP(rr, req)

	mockTokenParser.AssertExpectations(t)
}

func TestJwtAuthorizationHeaderToContext_NoBearerToken(t *testing.T) {
	mockTokenParser := new(OAuthMocks.TokenParser)
	mockDatabase := new(dataMocks.Database)
	mockAuthHelper := new(helpersMocks.AuthHelper)
	middleware := NewMiddlewareJwt(nil, mockTokenParser, mockDatabase, mockAuthHelper, nil)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value(constants.ContextKeyBearerToken)
		assert.Nil(t, token)
	})

	handler := middleware.JwtAuthorizationHeaderToContext()(nextHandler)
	handler.ServeHTTP(rr, req)
}

func TestJwtAuthorizationHeaderToContext_InvalidAuthorizationHeader(t *testing.T) {
	mockTokenParser := new(OAuthMocks.TokenParser)
	mockDatabase := new(dataMocks.Database)
	mockAuthHelper := new(helpersMocks.AuthHelper)
	middleware := NewMiddlewareJwt(nil, mockTokenParser, mockDatabase, mockAuthHelper, nil)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "NotBearer token")
	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value(constants.ContextKeyBearerToken)
		assert.Nil(t, token)
	})

	handler := middleware.JwtAuthorizationHeaderToContext()(nextHandler)
	handler.ServeHTTP(rr, req)
}

func TestRefreshToken_Success(t *testing.T) {
	mockTokenParser := new(OAuthMocks.TokenParser)
	mockDatabase := new(dataMocks.Database)
	mockAuthHelper := new(helpersMocks.AuthHelper)
	mockSessionStore := new(storeMocks.Store)
	mockHTTPClient := &mockHTTPClient{}
	middleware := NewMiddlewareJwt(mockSessionStore, mockTokenParser, mockDatabase, mockAuthHelper, mockHTTPClient)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	initialTokenResponse := &oauth.TokenResponse{
		AccessToken:  "oldaccesstoken",
		RefreshToken: "oldrefreshtoken",
	}

	session := &sessions.Session{
		Values: map[interface{}]interface{}{
			constants.SessionKeyJwt: *initialTokenResponse,
		},
	}

	mockSessionStore.On("Get", mock.Anything, constants.SessionName).Return(session, nil)
	mockSessionStore.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	aesEncryptionKey := "test_encryption_key_000000000000"
	clientSecretEncrypted, _ := encryption.EncryptText("encrypted_secret", []byte(aesEncryptionKey))
	mockDatabase.On("GetClientByClientIdentifier", mock.Anything, constants.AdminConsoleClientIdentifier).Return(&models.Client{
		ClientSecretEncrypted: clientSecretEncrypted,
	}, nil)

	settings := &models.Settings{
		Issuer:           "https://example.com",
		AESEncryptionKey: []byte(aesEncryptionKey),
	}
	ctx := context.WithValue(req.Context(), constants.ContextKeySettings, settings)
	req = req.WithContext(ctx)

	mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(`{
			"access_token": "newaccesstoken",
			"token_type": "Bearer",
			"expires_in": 3600,
			"refresh_token": "newrefreshtoken"
		}`)),
	}, nil)

	refreshed, err := middleware.refreshToken(rr, req, initialTokenResponse)

	assert.True(t, refreshed)
	assert.NoError(t, err)

	newTokenResponse, ok := session.Values[constants.SessionKeyJwt].(oauth.TokenResponse)
	assert.True(t, ok)
	assert.Equal(t, "newaccesstoken", newTokenResponse.AccessToken)
	assert.Equal(t, "newrefreshtoken", newTokenResponse.RefreshToken)

	mockSessionStore.AssertExpectations(t)
	mockDatabase.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}

func TestRefreshToken_NoRefreshToken(t *testing.T) {
	mockTokenParser := new(OAuthMocks.TokenParser)
	mockDatabase := new(dataMocks.Database)
	mockAuthHelper := new(helpersMocks.AuthHelper)
	mockSessionStore := new(storeMocks.Store)
	mockHTTPClient := &mockHTTPClient{}
	middleware := NewMiddlewareJwt(mockSessionStore, mockTokenParser, mockDatabase, mockAuthHelper, mockHTTPClient)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	tokenResponse := &oauth.TokenResponse{
		AccessToken: "oldaccesstoken",
		// No refresh token
	}
	refreshed, err := middleware.refreshToken(rr, req, tokenResponse)

	assert.False(t, refreshed)
	assert.NoError(t, err)
}

func TestRefreshToken_InvalidResponse(t *testing.T) {
	mockTokenParser := new(OAuthMocks.TokenParser)
	mockDatabase := new(dataMocks.Database)
	mockAuthHelper := new(helpersMocks.AuthHelper)
	mockSessionStore := new(storeMocks.Store)
	mockHTTPClient := &mockHTTPClient{}
	middleware := NewMiddlewareJwt(mockSessionStore, mockTokenParser, mockDatabase, mockAuthHelper, mockHTTPClient)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	tokenResponse := &oauth.TokenResponse{
		AccessToken:  "oldaccesstoken",
		RefreshToken: "oldrefreshtoken",
	}

	aesEncryptionKey := "test_encryption_key_000000000000"
	clientSecretEncrypted, _ := encryption.EncryptText("encrypted_secret", []byte(aesEncryptionKey))
	mockDatabase.On("GetClientByClientIdentifier", mock.Anything, constants.AdminConsoleClientIdentifier).Return(&models.Client{
		ClientSecretEncrypted: clientSecretEncrypted,
	}, nil)

	settings := &models.Settings{
		Issuer:           "https://example.com",
		AESEncryptionKey: []byte(aesEncryptionKey),
	}
	ctx := context.WithValue(req.Context(), constants.ContextKeySettings, settings)
	req = req.WithContext(ctx)

	mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(`{"error": "invalid_grant"}`)),
	}, nil)

	refreshed, err := middleware.refreshToken(rr, req, tokenResponse)

	assert.False(t, refreshed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error response from server")

	mockSessionStore.AssertExpectations(t)
	mockDatabase.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}
