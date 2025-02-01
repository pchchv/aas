package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pchchv/aas/src/constants"
	dataMocks "github.com/pchchv/aas/src/database/mocks"
	helpersMocks "github.com/pchchv/aas/src/helpers/mocks"
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
