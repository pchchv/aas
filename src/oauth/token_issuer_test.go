package oauth

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pchchv/aas/src/config"
	"github.com/pchchv/aas/src/constants"
	"github.com/pchchv/aas/src/database/mocks"
	"github.com/pchchv/aas/src/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateTokenResponseForAuthCode_FullOpenIDConnect(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	mockTokenParser := &TokenParser{}
	tokenIssuer := NewTokenIssuer(mockDB, mockTokenParser)
	settings := &models.Settings{
		Issuer:                                  "https://test-issuer.com",
		TokenExpirationInSeconds:                600,
		UserSessionIdleTimeoutInSeconds:         1200,
		UserSessionMaxLifetimeInSeconds:         2400,
		IncludeOpenIDConnectClaimsInAccessToken: true,
		RefreshTokenOfflineIdleTimeoutInSeconds: 1800,
		RefreshTokenOfflineMaxLifetimeInSeconds: 3600,
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeySettings, settings)
	now := time.Now().UTC()
	sub := uuid.New()
	sessionIdentifier := "test-session-123"
	config.Get().BaseURL = "http://localhost:8081"
	privateKeyBytes := getTestPrivateKey(t)
	publicKeyBytes := getTestPublicKey(t)
	code := &models.Code{
		Id:                1,
		ClientId:          1,
		UserId:            1,
		Scope:             "openid profile email address phone groups attributes offline_access",
		Nonce:             "test-nonce",
		AuthenticatedAt:   now.Add(-5 * time.Minute),
		SessionIdentifier: sessionIdentifier,
		AcrLevel:          "urn:goiabada:pwd:otp_mandatory",
		AuthMethods:       "pwd otp",
	}
	client := &models.Client{
		Id:                                      1,
		ClientIdentifier:                        "test-client",
		TokenExpirationInSeconds:                900,
		RefreshTokenOfflineIdleTimeoutInSeconds: 3600,
		RefreshTokenOfflineMaxLifetimeInSeconds: 7200,
	}
	user := &models.User{
		Id:                  1,
		UpdatedAt:           sql.NullTime{Time: time.Now().Add(-1 * time.Minute), Valid: true},
		Subject:             sub,
		Email:               "test@example.com",
		EmailVerified:       true,
		Username:            "testuser",
		GivenName:           "Test",
		MiddleName:          "Middle",
		FamilyName:          "User",
		Nickname:            "Testy",
		Website:             "https://test.com",
		Gender:              "male",
		BirthDate:           sql.NullTime{Time: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		ZoneInfo:            "Europe/London",
		Locale:              "en-GB",
		PhoneNumber:         "+1234567890",
		PhoneNumberVerified: true,
		AddressLine1:        "123 Test St",
		AddressLine2:        "apartment 1",
		AddressLocality:     "Test City",
		AddressRegion:       "Test Region",
		AddressPostalCode:   "12345",
		AddressCountry:      "Test Country",
		Groups: []models.Group{
			{GroupIdentifier: "group1", IncludeInIdToken: true, IncludeInAccessToken: true},
			{GroupIdentifier: "group2", IncludeInIdToken: true, IncludeInAccessToken: false},
			{GroupIdentifier: "group3", IncludeInIdToken: false, IncludeInAccessToken: true},
			{GroupIdentifier: "group4", IncludeInIdToken: true, IncludeInAccessToken: true},
		},
		Attributes: []models.UserAttribute{
			{Key: "attr1", Value: "value1", IncludeInIdToken: true, IncludeInAccessToken: true},
			{Key: "attr2", Value: "value2", IncludeInIdToken: true, IncludeInAccessToken: false},
			{Key: "attr3", Value: "value3", IncludeInIdToken: false, IncludeInAccessToken: true},
			{Key: "attr4", Value: "value4", IncludeInIdToken: true, IncludeInAccessToken: true},
		},
	}

	mockDB.On("CodeLoadClient", mock.Anything, code).Return(nil)
	code.Client = *client
	mockDB.On("CodeLoadUser", mock.Anything, code).Return(nil)
	code.User = *user
	mockDB.On("UserLoadGroups", mock.Anything, &code.User).Return(nil)
	mockDB.On("GroupsLoadAttributes", mock.Anything, code.User.Groups).Return(nil)
	mockDB.On("UserLoadAttributes", mock.Anything, &code.User).Return(nil)
	mockDB.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockDB.On("GetCurrentSigningKey", mock.Anything).Return(&models.KeyPair{
		KeyIdentifier: "test-key-id",
		PrivateKeyPEM: privateKeyBytes,
	}, nil)

	response, err := tokenIssuer.GenerateTokenResponseForAuthCode(ctx, code)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(900), response.ExpiresIn)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.IdToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "openid profile email address phone groups attributes offline_access authserver:userinfo", response.Scope)
	assert.Equal(t, int64(3600), response.RefreshExpiresIn)

	// validate Id token --------------------------------------------

	idClaims := verifyAndDecodeToken(t, response.IdToken, publicKeyBytes)
	assert.Equal(t, settings.Issuer, idClaims["iss"])
	assert.Equal(t, user.Subject.String(), idClaims["sub"])
	assert.Equal(t, client.ClientIdentifier, idClaims["aud"])
	assert.Equal(t, code.Nonce, idClaims["nonce"])
	assert.Equal(t, code.AcrLevel, idClaims["acr"])
	assert.Equal(t, code.AuthMethods, idClaims["amr"])
	assert.Equal(t, sessionIdentifier, idClaims["sid"])
	assert.Equal(t, "ID", idClaims["typ"])
	assertTimeClaimWithinRange(t, idClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, idClaims, "exp", 900*time.Second, "exp should be 900 seconds from now")
	assertTimeClaimWithinRange(t, idClaims, "updated_at", -60*time.Second, "updated_at should be 60 seconds ago")
	assertTimeClaimWithinRange(t, idClaims, "auth_time", -300*time.Second, "auth_time should be 300 seconds ago")
	assert.Contains(t, idClaims, "auth_time")
	authTimeUnix := idClaims["auth_time"].(float64)
	authTime := time.Unix(int64(authTimeUnix), 0)
	assert.Equal(t, now.Add(-300*time.Second).Unix(), authTime.Unix(), fmt.Sprintf("auth_time should be 300 seconds ago: %s", authTime))
	_, err = uuid.Parse(idClaims["jti"].(string))
	assert.NoError(t, err)
	assert.Equal(t, user.GetFullName(), idClaims["name"])
	assert.Equal(t, user.GivenName, idClaims["given_name"])
	assert.Equal(t, user.MiddleName, idClaims["middle_name"])
	assert.Equal(t, user.FamilyName, idClaims["family_name"])
	assert.Equal(t, user.Nickname, idClaims["nickname"])
	assert.Equal(t, user.Username, idClaims["preferred_username"])
	assert.Equal(t, "http://localhost:8081/account/profile", idClaims["profile"])
	assert.Equal(t, user.Website, idClaims["website"])
	assert.Equal(t, user.Gender, idClaims["gender"])
	assert.Equal(t, "1990-01-01", idClaims["birthdate"])
	assert.Equal(t, user.ZoneInfo, idClaims["zoneinfo"])
	assert.Equal(t, user.Locale, idClaims["locale"])
	assert.NotEmpty(t, idClaims["updated_at"])
	assert.Equal(t, user.Email, idClaims["email"])
	assert.Equal(t, user.EmailVerified, idClaims["email_verified"])
	assert.Equal(t, user.PhoneNumber, idClaims["phone_number"])
	assert.Equal(t, user.PhoneNumberVerified, idClaims["phone_number_verified"])
	address, ok := idClaims["address"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, user.AddressLine1+"\r\n"+user.AddressLine2, address["street_address"])
	assert.Equal(t, user.AddressLocality, address["locality"])
	assert.Equal(t, user.AddressRegion, address["region"])
	assert.Equal(t, user.AddressPostalCode, address["postal_code"])
	assert.Equal(t, user.AddressCountry, address["country"])
	assert.Equal(t, "123 Test St\r\napartment 1\r\nTest City\r\nTest Region\r\n12345\r\nTest Country", address["formatted"])
	groups, ok := idClaims["groups"].([]interface{})
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{"group1", "group2", "group4"}, groups)
	assert.Equal(t, 3, len(groups))
	attributes, ok := idClaims["attributes"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value1", attributes["attr1"])
	assert.Equal(t, "value2", attributes["attr2"])
	assert.Equal(t, "value4", attributes["attr4"])
	assert.Equal(t, 3, len(attributes))

	// validate Access token --------------------------------------------

	accessClaims := verifyAndDecodeToken(t, response.AccessToken, publicKeyBytes)
	assert.Equal(t, settings.Issuer, accessClaims["iss"])
	assert.Equal(t, user.Subject.String(), accessClaims["sub"])
	assert.Equal(t, "authserver", accessClaims["aud"])
	assert.Equal(t, code.Nonce, accessClaims["nonce"])
	assert.Equal(t, code.AcrLevel, accessClaims["acr"])
	assert.Equal(t, code.AuthMethods, accessClaims["amr"])
	assert.Equal(t, sessionIdentifier, accessClaims["sid"])
	assert.Equal(t, "Bearer", accessClaims["typ"])

	assertTimeClaimWithinRange(t, accessClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, accessClaims, "exp", 900*time.Second, "exp should be 900 seconds from now")
	assertTimeClaimWithinRange(t, accessClaims, "updated_at", -60*time.Second, "updated_at should be 60 seconds ago")
	assertTimeClaimWithinRange(t, accessClaims, "auth_time", -300*time.Second, "auth_time should be 300 seconds ago")

	_, err = uuid.Parse(accessClaims["jti"].(string))
	assert.NoError(t, err)

	assert.Equal(t, user.GetFullName(), accessClaims["name"])
	assert.Equal(t, user.GivenName, accessClaims["given_name"])
	assert.Equal(t, user.MiddleName, accessClaims["middle_name"])
	assert.Equal(t, user.FamilyName, accessClaims["family_name"])
	assert.Equal(t, user.Nickname, accessClaims["nickname"])
	assert.Equal(t, user.Username, accessClaims["preferred_username"])
	assert.Equal(t, "http://localhost:8081/account/profile", accessClaims["profile"])
	assert.Equal(t, user.Website, accessClaims["website"])
	assert.Equal(t, user.Gender, accessClaims["gender"])
	assert.Equal(t, "1990-01-01", accessClaims["birthdate"])
	assert.Equal(t, user.ZoneInfo, accessClaims["zoneinfo"])
	assert.Equal(t, user.Locale, accessClaims["locale"])
	assert.NotEmpty(t, accessClaims["updated_at"])
	assert.Equal(t, user.Email, accessClaims["email"])
	assert.Equal(t, user.EmailVerified, accessClaims["email_verified"])
	assert.Equal(t, user.PhoneNumber, accessClaims["phone_number"])
	assert.Equal(t, user.PhoneNumberVerified, accessClaims["phone_number_verified"])
	address, ok = accessClaims["address"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, user.AddressLine1+"\r\n"+user.AddressLine2, address["street_address"])
	assert.Equal(t, user.AddressLocality, address["locality"])
	assert.Equal(t, user.AddressRegion, address["region"])
	assert.Equal(t, user.AddressPostalCode, address["postal_code"])
	assert.Equal(t, user.AddressCountry, address["country"])
	assert.Equal(t, "123 Test St\r\napartment 1\r\nTest City\r\nTest Region\r\n12345\r\nTest Country", address["formatted"])
	groups, ok = accessClaims["groups"].([]interface{})
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{"group1", "group3", "group4"}, groups)
	assert.Equal(t, 3, len(groups))
	attributes, ok = accessClaims["attributes"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value1", attributes["attr1"])
	assert.Equal(t, "value3", attributes["attr3"])
	assert.Equal(t, "value4", attributes["attr4"])
	assert.Equal(t, 3, len(attributes))
	assert.Equal(t, "openid profile email address phone groups attributes offline_access authserver:userinfo", accessClaims["scope"])

	// validate Refresh token --------------------------------------------

	refreshClaims := verifyAndDecodeToken(t, response.RefreshToken, publicKeyBytes)
	assert.Equal(t, user.Subject.String(), refreshClaims["sub"])
	assert.Equal(t, "https://test-issuer.com", refreshClaims["aud"])
	assert.Equal(t, "https://test-issuer.com", refreshClaims["iss"])
	assert.Equal(t, "Offline", refreshClaims["typ"])
	assert.Equal(t, "openid profile email address phone groups attributes offline_access authserver:userinfo", refreshClaims["scope"])

	assertTimeClaimWithinRange(t, refreshClaims, "exp", 3600*time.Second, "exp should be 3600 seconds from now")
	assertTimeClaimWithinRange(t, refreshClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, refreshClaims, "offline_access_max_lifetime", 7200*time.Second, "offline_access_max_lifetime should be 7200 seconds from now")

	_, err = uuid.Parse(refreshClaims["jti"].(string))
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestGenerateTokenResponseForAuthCode_MinimalScope(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	mockTokenParser := &TokenParser{}
	tokenIssuer := NewTokenIssuer(mockDB, mockTokenParser)
	settings := &models.Settings{
		Issuer:                                  "https://test-issuer.com",
		TokenExpirationInSeconds:                600,
		UserSessionIdleTimeoutInSeconds:         1200,
		UserSessionMaxLifetimeInSeconds:         2400,
		IncludeOpenIDConnectClaimsInAccessToken: false,
		RefreshTokenOfflineIdleTimeoutInSeconds: 1800,
		RefreshTokenOfflineMaxLifetimeInSeconds: 3600,
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeySettings, settings)
	now := time.Now().UTC()
	sub := uuid.New()
	sessionIdentifier := "test-session-123"
	config.Get().BaseURL = "http://localhost:8081"

	privateKeyBytes := getTestPrivateKey(t)
	publicKeyBytes := getTestPublicKey(t)
	code := &models.Code{
		Id:                2,
		ClientId:          2,
		UserId:            2,
		Scope:             "openid",
		Nonce:             "minimal-nonce",
		AuthenticatedAt:   now.Add(-120 * time.Second), // Authenticated 2 minutes ago
		SessionIdentifier: sessionIdentifier,
		AcrLevel:          "urn:goiabada:pwd",
		AuthMethods:       "pwd",
	}
	client := &models.Client{
		Id:               2,
		ClientIdentifier: "minimal-client",
	}
	user := &models.User{
		Id:      2,
		Subject: sub,
		Email:   "minimal@example.com",
	}

	mockDB.On("CodeLoadClient", mock.Anything, code).Return(nil)
	code.Client = *client
	mockDB.On("CodeLoadUser", mock.Anything, code).Return(nil)
	code.User = *user
	mockDB.On("UserLoadGroups", mock.Anything, &code.User).Return(nil)
	mockDB.On("GroupsLoadAttributes", mock.Anything, code.User.Groups).Return(nil)
	mockDB.On("UserLoadAttributes", mock.Anything, &code.User).Return(nil)
	mockDB.On("GetUserSessionBySessionIdentifier", mock.Anything, sessionIdentifier).Return(&models.UserSession{
		Id:           1,
		UserId:       1,
		Started:      now.Add(-30 * time.Minute),
		LastAccessed: now.Add(-5 * time.Minute),
	}, nil)
	mockDB.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockDB.On("GetCurrentSigningKey", mock.Anything).Return(&models.KeyPair{
		KeyIdentifier: "test-key-id",
		PrivateKeyPEM: privateKeyBytes,
	}, nil)

	response, err := tokenIssuer.GenerateTokenResponseForAuthCode(ctx, code)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(600), response.ExpiresIn)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.IdToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "openid authserver:userinfo", response.Scope)
	assert.InDelta(t, int64(600), response.RefreshExpiresIn, 1)

	// validate Id token --------------------------------------------

	idClaims := verifyAndDecodeToken(t, response.IdToken, publicKeyBytes)
	assert.Equal(t, "https://test-issuer.com", idClaims["iss"])
	assert.Equal(t, user.Subject.String(), idClaims["sub"])
	assert.Equal(t, client.ClientIdentifier, idClaims["aud"])
	assert.Equal(t, code.Nonce, idClaims["nonce"])
	assert.Equal(t, code.AcrLevel, idClaims["acr"])
	assert.Equal(t, code.AuthMethods, idClaims["amr"])
	assert.Equal(t, sessionIdentifier, idClaims["sid"])

	assertTimeClaimWithinRange(t, idClaims, "auth_time", -120*time.Second, "auth_time should be 2 minutes ago")
	assertTimeClaimWithinRange(t, idClaims, "exp", 600*time.Second, "exp should be 10 minutes in the future")
	assertTimeClaimWithinRange(t, idClaims, "iat", 0, "iat should be now")

	_, err = uuid.Parse(idClaims["jti"].(string))
	assert.NoError(t, err)

	assert.Equal(t, "ID", idClaims["typ"])

	// validate Access token --------------------------------------------

	accessClaims := verifyAndDecodeToken(t, response.AccessToken, publicKeyBytes)
	assert.Equal(t, "https://test-issuer.com", accessClaims["iss"])
	assert.Equal(t, user.Subject.String(), accessClaims["sub"])
	assert.Equal(t, code.Nonce, accessClaims["nonce"])
	assert.Equal(t, code.AcrLevel, accessClaims["acr"])
	assert.Equal(t, code.AuthMethods, accessClaims["amr"])
	assert.Equal(t, sessionIdentifier, accessClaims["sid"])
	assert.Equal(t, "Bearer", accessClaims["typ"])
	assert.Equal(t, "openid authserver:userinfo", accessClaims["scope"])

	assertTimeClaimWithinRange(t, accessClaims, "auth_time", -120*time.Second, "auth_time should be 2 minutes ago")
	assertTimeClaimWithinRange(t, accessClaims, "exp", 600*time.Second, "exp should be 10 minutes in the future")
	assertTimeClaimWithinRange(t, accessClaims, "iat", 0, "iat should be now")

	_, err = uuid.Parse(accessClaims["jti"].(string))
	assert.NoError(t, err)

	// validate Refresh token --------------------------------------------

	refreshClaims := verifyAndDecodeToken(t, response.RefreshToken, publicKeyBytes)
	assert.Equal(t, user.Subject.String(), refreshClaims["sub"])
	assert.Equal(t, "https://test-issuer.com", refreshClaims["aud"])
	assert.Equal(t, "Refresh", refreshClaims["typ"])
	assert.Equal(t, sessionIdentifier, refreshClaims["sid"])
	assert.Equal(t, "openid authserver:userinfo", refreshClaims["scope"])

	assertTimeClaimWithinRange(t, refreshClaims, "exp", 600*time.Second, "exp should be 10 minutes in the future")
	assertTimeClaimWithinRange(t, refreshClaims, "iat", 0, "iat should be now")

	_, err = uuid.Parse(refreshClaims["jti"].(string))
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestGenerateTokenResponseForAuthCode_ClientOverrideAndMixedScopes(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	mockTokenParser := &TokenParser{}
	tokenIssuer := NewTokenIssuer(mockDB, mockTokenParser)

	settings := &models.Settings{
		Issuer:                                  "https://test-issuer.com",
		TokenExpirationInSeconds:                600,
		UserSessionIdleTimeoutInSeconds:         1200,
		UserSessionMaxLifetimeInSeconds:         2400,
		IncludeOpenIDConnectClaimsInAccessToken: false,
		RefreshTokenOfflineIdleTimeoutInSeconds: 1800,
		RefreshTokenOfflineMaxLifetimeInSeconds: 3600,
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeySettings, settings)

	now := time.Now().UTC()
	sub := uuid.New()
	sessionIdentifier := "test-session-123"
	config.Get().BaseURL = "http://localhost:8081"

	privateKeyBytes := getTestPrivateKey(t)
	publicKeyBytes := getTestPublicKey(t)

	code := &models.Code{
		Id:                3,
		ClientId:          3,
		UserId:            3,
		Scope:             "openid profile email authserver:userinfo resource1:read resource2:write",
		Nonce:             "mixed-nonce",
		AuthenticatedAt:   now.Add(-60 * time.Second),
		SessionIdentifier: sessionIdentifier,
		AcrLevel:          "urn:goiabada:pwd:otp_ifpossible",
		AuthMethods:       "pwd otp",
	}
	client := &models.Client{
		Id:                                      3,
		ClientIdentifier:                        "mixed-client",
		TokenExpirationInSeconds:                1500,
		RefreshTokenOfflineIdleTimeoutInSeconds: 2400,
		RefreshTokenOfflineMaxLifetimeInSeconds: 4800,
		IncludeOpenIDConnectClaimsInAccessToken: "on",
	}
	user := &models.User{
		Id:            3,
		Subject:       sub,
		Email:         "mixed@example.com",
		EmailVerified: true,
		Username:      "mixeduser",
		GivenName:     "Mixed",
		FamilyName:    "User",
		UpdatedAt:     sql.NullTime{Time: now.Add(-24 * time.Hour), Valid: true},
		Groups: []models.Group{
			{GroupIdentifier: "group1", IncludeInIdToken: true, IncludeInAccessToken: true},
			{GroupIdentifier: "group2", IncludeInIdToken: false, IncludeInAccessToken: true},
		},
		Attributes: []models.UserAttribute{
			{Key: "attr1", Value: "value1", IncludeInIdToken: true, IncludeInAccessToken: true},
			{Key: "attr2", Value: "value2", IncludeInIdToken: true, IncludeInAccessToken: false},
		},
	}

	mockDB.On("CodeLoadClient", mock.Anything, code).Return(nil)
	code.Client = *client
	mockDB.On("CodeLoadUser", mock.Anything, code).Return(nil)
	code.User = *user
	mockDB.On("UserLoadGroups", mock.Anything, &code.User).Return(nil)
	mockDB.On("GroupsLoadAttributes", mock.Anything, code.User.Groups).Return(nil)
	mockDB.On("UserLoadAttributes", mock.Anything, &code.User).Return(nil)
	mockDB.On("GetUserSessionBySessionIdentifier", mock.Anything, sessionIdentifier).Return(&models.UserSession{
		Id:           1,
		UserId:       3,
		Started:      now.Add(-30 * time.Minute),
		LastAccessed: now.Add(-5 * time.Minute),
	}, nil)
	mockDB.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockDB.On("GetCurrentSigningKey", mock.Anything).Return(&models.KeyPair{
		KeyIdentifier: "test-key-id",
		PrivateKeyPEM: privateKeyBytes,
	}, nil)

	response, err := tokenIssuer.GenerateTokenResponseForAuthCode(ctx, code)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(1500), response.ExpiresIn)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.IdToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "openid profile email authserver:userinfo resource1:read resource2:write", response.Scope)
	assert.InDelta(t, int64(600), response.RefreshExpiresIn, 1)

	// validate Id token --------------------------------------------

	idClaims := verifyAndDecodeToken(t, response.IdToken, publicKeyBytes)
	assert.Equal(t, settings.Issuer, idClaims["iss"])
	assert.Equal(t, user.Subject.String(), idClaims["sub"])
	assert.Equal(t, client.ClientIdentifier, idClaims["aud"])
	assert.Equal(t, code.Nonce, idClaims["nonce"])
	assert.Equal(t, code.AcrLevel, idClaims["acr"])
	assert.Equal(t, code.AuthMethods, idClaims["amr"])
	assert.Equal(t, sessionIdentifier, idClaims["sid"])
	assert.Equal(t, "ID", idClaims["typ"])

	assertTimeClaimWithinRange(t, idClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, idClaims, "exp", 1500*time.Second, "exp should be 1500 seconds from now")
	assertTimeClaimWithinRange(t, idClaims, "auth_time", -60*time.Second, "auth_time should be 60 seconds ago")
	assertTimeClaimWithinRange(t, idClaims, "updated_at", -24*time.Hour, "updated_at should be 24 hours ago")

	_, err = uuid.Parse(idClaims["jti"].(string))
	assert.NoError(t, err)

	assert.Equal(t, user.Email, idClaims["email"])
	assert.Equal(t, user.EmailVerified, idClaims["email_verified"])
	assert.Equal(t, user.Username, idClaims["preferred_username"])
	assert.Equal(t, user.GivenName, idClaims["given_name"])
	assert.Equal(t, user.FamilyName, idClaims["family_name"])
	assert.Equal(t, user.GetFullName(), idClaims["name"])
	assert.Equal(t, "http://localhost:8081/account/profile", idClaims["profile"])

	assert.NotContains(t, idClaims, "groups")
	assert.NotContains(t, idClaims, "attributes")

	// validate Access token --------------------------------------------

	accessClaims := verifyAndDecodeToken(t, response.AccessToken, publicKeyBytes)
	assert.Equal(t, settings.Issuer, accessClaims["iss"])
	assert.Equal(t, user.Subject.String(), accessClaims["sub"])
	assert.Equal(t, []interface{}{"authserver", "resource1", "resource2"}, accessClaims["aud"])
	assert.Equal(t, code.Nonce, accessClaims["nonce"])
	assert.Equal(t, code.AcrLevel, accessClaims["acr"])
	assert.Equal(t, code.AuthMethods, accessClaims["amr"])
	assert.Equal(t, sessionIdentifier, accessClaims["sid"])
	assert.Equal(t, "Bearer", accessClaims["typ"])

	assertTimeClaimWithinRange(t, accessClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, accessClaims, "exp", 1500*time.Second, "exp should be 1500 seconds from now")
	assertTimeClaimWithinRange(t, accessClaims, "auth_time", -60*time.Second, "auth_time should be 60 seconds ago")
	assertTimeClaimWithinRange(t, accessClaims, "updated_at", -24*time.Hour, "updated_at should be 24 hours ago")

	_, err = uuid.Parse(accessClaims["jti"].(string))
	assert.NoError(t, err)

	assert.Equal(t, user.Email, accessClaims["email"])
	assert.Equal(t, user.EmailVerified, accessClaims["email_verified"])
	assert.Equal(t, user.Username, accessClaims["preferred_username"])
	assert.Equal(t, user.GivenName, accessClaims["given_name"])
	assert.Equal(t, user.FamilyName, accessClaims["family_name"])
	assert.Equal(t, user.GetFullName(), accessClaims["name"])
	assert.Equal(t, "http://localhost:8081/account/profile", accessClaims["profile"])

	assert.NotContains(t, accessClaims, "groups")
	assert.NotContains(t, accessClaims, "attributes")
	assert.Equal(t, "openid profile email authserver:userinfo resource1:read resource2:write", accessClaims["scope"])

	// validate Refresh token --------------------------------------------

	refreshClaims := verifyAndDecodeToken(t, response.RefreshToken, publicKeyBytes)
	assert.Equal(t, user.Subject.String(), refreshClaims["sub"])
	assert.Equal(t, settings.Issuer, refreshClaims["aud"])
	assert.Equal(t, settings.Issuer, refreshClaims["iss"])
	assert.Equal(t, "Refresh", refreshClaims["typ"])
	assert.Equal(t, sessionIdentifier, refreshClaims["sid"])
	assert.Equal(t, "openid profile email authserver:userinfo resource1:read resource2:write", refreshClaims["scope"])

	assertTimeClaimWithinRange(t, refreshClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, refreshClaims, "exp", 600*time.Second, "exp should be 600 seconds from now")

	_, err = uuid.Parse(refreshClaims["jti"].(string))
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestGenerateTokenResponseForAuthCode_ClientOverrideAndCustomScope(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	mockTokenParser := &TokenParser{}
	tokenIssuer := NewTokenIssuer(mockDB, mockTokenParser)

	settings := &models.Settings{
		Issuer:                                  "https://test-issuer.com",
		TokenExpirationInSeconds:                600,
		UserSessionIdleTimeoutInSeconds:         1200,
		UserSessionMaxLifetimeInSeconds:         2400,
		IncludeOpenIDConnectClaimsInAccessToken: false,
		RefreshTokenOfflineIdleTimeoutInSeconds: 1800,
		RefreshTokenOfflineMaxLifetimeInSeconds: 3600,
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeySettings, settings)

	now := time.Now().UTC()
	sub := uuid.New()
	sessionIdentifier := "test-session-123"
	config.Get().BaseURL = "http://localhost:8081"

	privateKeyBytes := getTestPrivateKey(t)
	publicKeyBytes := getTestPublicKey(t)

	code := &models.Code{
		Id:                4,
		ClientId:          4,
		UserId:            4,
		Scope:             "resource1:read resource2:write offline_access",
		Nonce:             "custom-nonce",
		AuthenticatedAt:   now.Add(-30 * time.Second),
		SessionIdentifier: sessionIdentifier,
		AcrLevel:          "urn:goiabada:pwd",
		AuthMethods:       "pwd",
	}
	client := &models.Client{
		Id:                                      4,
		ClientIdentifier:                        "custom-client",
		TokenExpirationInSeconds:                1200,
		RefreshTokenOfflineIdleTimeoutInSeconds: 3000,
		RefreshTokenOfflineMaxLifetimeInSeconds: 6000,
		IncludeOpenIDConnectClaimsInAccessToken: "off",
	}
	user := &models.User{
		Id:      4,
		Subject: sub,
		Email:   "custom@example.com",
	}

	mockDB.On("CodeLoadClient", mock.Anything, code).Return(nil)
	code.Client = *client
	mockDB.On("CodeLoadUser", mock.Anything, code).Return(nil)
	code.User = *user
	mockDB.On("UserLoadGroups", mock.Anything, &code.User).Return(nil)
	mockDB.On("GroupsLoadAttributes", mock.Anything, code.User.Groups).Return(nil)
	mockDB.On("UserLoadAttributes", mock.Anything, &code.User).Return(nil)
	mockDB.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockDB.On("GetCurrentSigningKey", mock.Anything).Return(&models.KeyPair{
		KeyIdentifier: "test-key-id",
		PrivateKeyPEM: privateKeyBytes,
	}, nil)

	response, err := tokenIssuer.GenerateTokenResponseForAuthCode(ctx, code)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(1200), response.ExpiresIn)
	assert.NotEmpty(t, response.AccessToken)
	assert.Empty(t, response.IdToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "resource1:read resource2:write offline_access", response.Scope)
	assert.Equal(t, int64(3000), response.RefreshExpiresIn)

	accessClaims := verifyAndDecodeToken(t, response.AccessToken, publicKeyBytes)
	assert.Equal(t, settings.Issuer, accessClaims["iss"])
	assert.Equal(t, user.Subject.String(), accessClaims["sub"])
	assert.Equal(t, []interface{}{"resource1", "resource2"}, accessClaims["aud"])
	assert.Equal(t, code.Nonce, accessClaims["nonce"])
	assert.Equal(t, code.AcrLevel, accessClaims["acr"])
	assert.Equal(t, code.AuthMethods, accessClaims["amr"])
	assert.Equal(t, sessionIdentifier, accessClaims["sid"])
	assert.Equal(t, "Bearer", accessClaims["typ"])

	assertTimeClaimWithinRange(t, accessClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, accessClaims, "exp", 1200*time.Second, "exp should be 1200 seconds from now")
	assertTimeClaimWithinRange(t, accessClaims, "auth_time", -30*time.Second, "auth_time should be 30 seconds ago")

	_, err = uuid.Parse(accessClaims["jti"].(string))
	assert.NoError(t, err)

	assert.Equal(t, "resource1:read resource2:write offline_access", accessClaims["scope"])
	assert.NotContains(t, accessClaims, "email")
	assert.NotContains(t, accessClaims, "name")

	refreshClaims := verifyAndDecodeToken(t, response.RefreshToken, publicKeyBytes)
	assert.Equal(t, user.Subject.String(), refreshClaims["sub"])
	assert.Equal(t, settings.Issuer, refreshClaims["aud"])
	assert.Equal(t, settings.Issuer, refreshClaims["iss"])
	assert.Equal(t, "Offline", refreshClaims["typ"])
	assert.Equal(t, "resource1:read resource2:write offline_access", refreshClaims["scope"])

	assertTimeClaimWithinRange(t, refreshClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, refreshClaims, "exp", 3000*time.Second, "exp should be 3000 seconds from now")
	assertTimeClaimWithinRange(t, refreshClaims, "offline_access_max_lifetime", 6000*time.Second, "offline_access_max_lifetime should be 6000 seconds from now")

	_, err = uuid.Parse(refreshClaims["jti"].(string))
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestGenerateTokenResponseForAuthCode_CustomScope(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	mockTokenParser := &TokenParser{}
	tokenIssuer := NewTokenIssuer(mockDB, mockTokenParser)

	settings := &models.Settings{
		Issuer:                                  "https://test-issuer.com",
		TokenExpirationInSeconds:                600,
		UserSessionIdleTimeoutInSeconds:         1200,
		UserSessionMaxLifetimeInSeconds:         2400,
		IncludeOpenIDConnectClaimsInAccessToken: false,
		RefreshTokenOfflineIdleTimeoutInSeconds: 1800,
		RefreshTokenOfflineMaxLifetimeInSeconds: 3600,
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeySettings, settings)

	now := time.Now().UTC()
	sub := uuid.New()
	sessionIdentifier := "test-session-123"
	config.Get().BaseURL = "http://localhost:8081"

	privateKeyBytes := getTestPrivateKey(t)
	publicKeyBytes := getTestPublicKey(t)

	code := &models.Code{
		Id:                5,
		ClientId:          5,
		UserId:            5,
		Scope:             "resource1:read",
		Nonce:             "custom-nonce",
		AuthenticatedAt:   now.Add(-30 * time.Second),
		SessionIdentifier: sessionIdentifier,
		AcrLevel:          "urn:goiabada:pwd",
		AuthMethods:       "pwd",
	}
	client := &models.Client{
		Id:               5,
		ClientIdentifier: "custom-scope-client",
	}
	user := &models.User{
		Id:      5,
		Subject: sub,
		Email:   "custom@example.com",
	}

	mockDB.On("CodeLoadClient", mock.Anything, code).Return(nil)
	code.Client = *client
	mockDB.On("CodeLoadUser", mock.Anything, code).Return(nil)
	code.User = *user
	mockDB.On("UserLoadGroups", mock.Anything, &code.User).Return(nil)
	mockDB.On("GroupsLoadAttributes", mock.Anything, code.User.Groups).Return(nil)
	mockDB.On("UserLoadAttributes", mock.Anything, &code.User).Return(nil)
	mockDB.On("GetUserSessionBySessionIdentifier", mock.Anything, sessionIdentifier).Return(&models.UserSession{
		Id:           1,
		UserId:       5,
		Started:      now.Add(-30 * time.Minute),
		LastAccessed: now.Add(-5 * time.Minute),
	}, nil)
	mockDB.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockDB.On("GetCurrentSigningKey", mock.Anything).Return(&models.KeyPair{
		KeyIdentifier: "test-key-id",
		PrivateKeyPEM: privateKeyBytes,
	}, nil)

	response, err := tokenIssuer.GenerateTokenResponseForAuthCode(ctx, code)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(600), response.ExpiresIn)
	assert.NotEmpty(t, response.AccessToken)
	assert.Empty(t, response.IdToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "resource1:read", response.Scope)
	assert.InDelta(t, int64(600), response.RefreshExpiresIn, 1)

	accessClaims := verifyAndDecodeToken(t, response.AccessToken, publicKeyBytes)
	assert.Equal(t, settings.Issuer, accessClaims["iss"])
	assert.Equal(t, user.Subject.String(), accessClaims["sub"])
	assert.Equal(t, "resource1", accessClaims["aud"])
	assert.Equal(t, code.Nonce, accessClaims["nonce"])
	assert.Equal(t, code.AcrLevel, accessClaims["acr"])
	assert.Equal(t, code.AuthMethods, accessClaims["amr"])
	assert.Equal(t, sessionIdentifier, accessClaims["sid"])
	assert.Equal(t, "Bearer", accessClaims["typ"])

	assertTimeClaimWithinRange(t, accessClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, accessClaims, "exp", 600*time.Second, "exp should be 600 seconds from now")
	assertTimeClaimWithinRange(t, accessClaims, "auth_time", -30*time.Second, "auth_time should be 30 seconds ago")

	_, err = uuid.Parse(accessClaims["jti"].(string))
	assert.NoError(t, err)

	assert.Equal(t, "resource1:read", accessClaims["scope"])
	assert.NotContains(t, accessClaims, "email")
	assert.NotContains(t, accessClaims, "name")

	refreshClaims := verifyAndDecodeToken(t, response.RefreshToken, publicKeyBytes)
	assert.Equal(t, user.Subject.String(), refreshClaims["sub"])
	assert.Equal(t, settings.Issuer, refreshClaims["aud"])
	assert.Equal(t, settings.Issuer, refreshClaims["iss"])
	assert.Equal(t, "Refresh", refreshClaims["typ"])
	assert.Equal(t, "resource1:read", refreshClaims["scope"])

	assertTimeClaimWithinRange(t, refreshClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, refreshClaims, "exp", 600*time.Second, "exp should be 600 seconds from now")

	_, err = uuid.Parse(refreshClaims["jti"].(string))
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestGenerateTokenResponseForClientCred(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	mockTokenParser := &TokenParser{}
	tokenIssuer := NewTokenIssuer(mockDB, mockTokenParser)

	settings := &models.Settings{
		Issuer:                   "https://test-issuer.com",
		TokenExpirationInSeconds: 3600,
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeySettings, settings)

	privateKeyBytes := getTestPrivateKey(t)
	publicKeyBytes := getTestPublicKey(t)

	tests := []struct {
		name           string
		client         *models.Client
		scope          string
		expectedScopes []string
		expectedAud    interface{}
	}{
		{
			name: "Single custom scope",
			client: &models.Client{
				Id:               1,
				ClientIdentifier: "test-client-1",
			},
			scope:          "resource1:read",
			expectedScopes: []string{"resource1:read"},
			expectedAud:    "resource1",
		},
		{
			name: "Multiple custom scopes",
			client: &models.Client{
				Id:               2,
				ClientIdentifier: "test-client-2",
			},
			scope:          "resource1:read resource2:write",
			expectedScopes: []string{"resource1:read", "resource2:write"},
			expectedAud:    []interface{}{"resource1", "resource2"},
		},
		{
			name: "Custom scopes with OIDC scopes (should be ignored)",
			client: &models.Client{
				Id:               3,
				ClientIdentifier: "test-client-3",
			},
			scope:          "resource1:read openid profile",
			expectedScopes: []string{"resource1:read"},
			expectedAud:    "resource1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB.On("GetCurrentSigningKey", mock.Anything).Return(&models.KeyPair{
				KeyIdentifier: "test-key-id",
				PrivateKeyPEM: privateKeyBytes,
			}, nil)

			response, err := tokenIssuer.GenerateTokenResponseForClientCred(ctx, tt.client, tt.scope)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, "Bearer", response.TokenType)
			assert.Equal(t, int64(3600), response.ExpiresIn)
			assert.NotEmpty(t, response.AccessToken)
			assert.Empty(t, response.IdToken)
			assert.Empty(t, response.RefreshToken)
			assert.Equal(t, tt.scope, response.Scope)

			claims := verifyAndDecodeToken(t, response.AccessToken, publicKeyBytes)

			assert.Equal(t, settings.Issuer, claims["iss"])
			assert.Equal(t, tt.client.ClientIdentifier, claims["sub"])
			assert.Equal(t, tt.expectedAud, claims["aud"])
			assert.Equal(t, "Bearer", claims["typ"])
			assert.Equal(t, tt.scope, claims["scope"])

			assertTimeClaimWithinRange(t, claims, "iat", 0*time.Second, "iat should be now")
			assertTimeClaimWithinRange(t, claims, "exp", 3600*time.Second, "exp should be 3600 seconds from now")

			_, err = uuid.Parse(claims["jti"].(string))
			assert.NoError(t, err)

			mockDB.AssertExpectations(t)
		})
	}
}

func TestGenerateTokenResponseForClientCred_InvalidScope(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	mockTokenParser := &TokenParser{}
	tokenIssuer := NewTokenIssuer(mockDB, mockTokenParser)

	settings := &models.Settings{
		Issuer:                   "https://test-issuer.com",
		TokenExpirationInSeconds: 3600,
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeySettings, settings)

	client := &models.Client{
		Id:               4,
		ClientIdentifier: "test-client-4",
	}

	privateKeyBytes := getTestPrivateKey(t)

	mockDB.On("GetCurrentSigningKey", mock.Anything).Return(&models.KeyPair{
		KeyIdentifier: "test-key-id",
		PrivateKeyPEM: privateKeyBytes,
	}, nil)

	response, err := tokenIssuer.GenerateTokenResponseForClientCred(ctx, client, "invalid-scope")

	if err == nil {
		t.Error("Expected an error, but got nil")
		if response != nil {
			t.Errorf("Unexpected response: %+v", response)
		}
	} else {
		assert.Contains(t, err.Error(), "invalid scope: invalid-scope")
	}

	mockDB.AssertExpectations(t)
}

func TestGenerateTokenResponseForRefresh(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	mockTokenParser := &TokenParser{}
	tokenIssuer := NewTokenIssuer(mockDB, mockTokenParser)

	settings := &models.Settings{
		Issuer:                                  "https://test-issuer.com",
		TokenExpirationInSeconds:                600,
		UserSessionIdleTimeoutInSeconds:         1200, // 20 minutes
		UserSessionMaxLifetimeInSeconds:         2400, // 40 minutes
		IncludeOpenIDConnectClaimsInAccessToken: true,
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeySettings, settings)

	now := time.Now().UTC()
	sub := uuid.New()
	sessionIdentifier := "test-session-123"
	config.Get().BaseURL = "http://localhost:8081"

	privateKeyBytes := getTestPrivateKey(t)
	publicKeyBytes := getTestPublicKey(t)

	code := &models.Code{
		Id:                1,
		ClientId:          1,
		UserId:            1,
		Scope:             "openid profile resource1:read",
		Nonce:             "test-nonce",
		AuthenticatedAt:   now.Add(-5 * time.Minute),
		SessionIdentifier: sessionIdentifier,
		AcrLevel:          "urn:goiabada:pwd",
		AuthMethods:       "pwd",
	}
	client := &models.Client{
		Id:                       1,
		ClientIdentifier:         "test-client",
		TokenExpirationInSeconds: 900,
	}
	user := &models.User{
		Id:            1,
		Subject:       sub,
		Email:         "test@example.com",
		EmailVerified: true,
		Username:      "testuser",
		GivenName:     "Test",
		FamilyName:    "User",
		UpdatedAt:     sql.NullTime{Time: now.Add(-1 * time.Hour), Valid: true},
	}

	refreshToken := &models.RefreshToken{
		Id:                   1,
		RefreshTokenJti:      "existing-jti",
		FirstRefreshTokenJti: "first-jti",
		MaxLifetime:          sql.NullTime{Time: now.Add(24 * time.Hour), Valid: true},
	}

	refreshTokenInfo := &Jwt{
		Claims: jwt.MapClaims{
			"jti":    "existing-jti",
			"scope":  "openid profile resource1:read",
			"exp":    now.Add(1 * time.Hour).Unix(),
			"iat":    now.Add(-1 * time.Hour).Unix(),
			"iss":    "https://test-issuer.com",
			"aud":    "https://test-issuer.com",
			"sub":    sub.String(),
			"typ":    "Refresh",
			"sid":    sessionIdentifier,
			"client": client.ClientIdentifier,
		},
	}

	mockDB.On("CodeLoadClient", mock.Anything, code).Return(nil)
	code.Client = *client
	mockDB.On("CodeLoadUser", mock.Anything, code).Return(nil)
	code.User = *user
	mockDB.On("UserLoadGroups", mock.Anything, &code.User).Return(nil)
	mockDB.On("GroupsLoadAttributes", mock.Anything, code.User.Groups).Return(nil)
	mockDB.On("UserLoadAttributes", mock.Anything, &code.User).Return(nil)
	var capturedRefreshToken *models.RefreshToken
	mockDB.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).
		Run(func(args mock.Arguments) {
			capturedRefreshToken = args.Get(1).(*models.RefreshToken)
		}).
		Return(nil)
	mockDB.On("GetCurrentSigningKey", mock.Anything).Return(&models.KeyPair{
		KeyIdentifier: "test-key-id",
		PrivateKeyPEM: privateKeyBytes,
	}, nil)
	// Add the missing mock expectation
	mockDB.On("GetUserSessionBySessionIdentifier", mock.Anything, sessionIdentifier).Return(&models.UserSession{
		Id:           1,
		UserId:       1,
		Started:      now.Add(-30 * time.Minute),
		LastAccessed: now.Add(-5 * time.Minute),
	}, nil)

	input := &GenerateTokenForRefreshInput{
		Code:             code,
		ScopeRequested:   "openid profile resource1:read",
		RefreshToken:     refreshToken,
		RefreshTokenInfo: refreshTokenInfo,
	}

	response, err := tokenIssuer.GenerateTokenResponseForRefresh(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(900), response.ExpiresIn) // client override
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.IdToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "openid profile resource1:read authserver:userinfo", response.Scope)
	assert.InDelta(t, int64(600), response.RefreshExpiresIn, 1) // remaining time based on session max lifetime

	// validate Id token --------------------------------------------

	idClaims := verifyAndDecodeToken(t, response.IdToken, publicKeyBytes)
	assert.Equal(t, settings.Issuer, idClaims["iss"])
	assert.Equal(t, user.Subject.String(), idClaims["sub"])
	assert.Equal(t, client.ClientIdentifier, idClaims["aud"])
	assert.Equal(t, code.Nonce, idClaims["nonce"])
	assert.Equal(t, code.AcrLevel, idClaims["acr"])
	assert.Equal(t, code.AuthMethods, idClaims["amr"])
	assert.Equal(t, sessionIdentifier, idClaims["sid"])
	assert.Equal(t, "ID", idClaims["typ"])
	assertTimeClaimWithinRange(t, idClaims, "auth_time", -300*time.Second, "auth_time should be 300 seconds ago")
	assertTimeClaimWithinRange(t, idClaims, "exp", 900*time.Second, "exp should be 900 seconds from now")
	assertTimeClaimWithinRange(t, idClaims, "iat", 0*time.Second, "iat should be now")
	assert.Equal(t, user.FamilyName, idClaims["family_name"])
	assert.Equal(t, user.GivenName, idClaims["given_name"])
	assert.Equal(t, user.GetFullName(), idClaims["name"])
	assert.Equal(t, user.Username, idClaims["preferred_username"])
	assert.Equal(t, fmt.Sprintf("%v/account/profile", config.Get().BaseURL), idClaims["profile"])
	_, err = uuid.Parse(idClaims["jti"].(string))
	assert.NoError(t, err)
	assertTimeClaimWithinRange(t, idClaims, "updated_at", -1*time.Hour, "updated_at should be 1 hour ago")

	// validate Access token --------------------------------------------

	accessClaims := verifyAndDecodeToken(t, response.AccessToken, publicKeyBytes)
	assert.Equal(t, settings.Issuer, accessClaims["iss"])
	assert.Equal(t, user.Subject.String(), accessClaims["sub"])
	assert.ElementsMatch(t, []string{constants.AuthServerResourceIdentifier, "resource1"}, accessClaims["aud"])
	assert.Equal(t, code.Nonce, accessClaims["nonce"])
	assert.Equal(t, code.AcrLevel, accessClaims["acr"])
	assert.Equal(t, code.AuthMethods, accessClaims["amr"])
	assert.Equal(t, sessionIdentifier, accessClaims["sid"])
	assert.Equal(t, "Bearer", accessClaims["typ"])
	assert.Equal(t, user.FamilyName, accessClaims["family_name"])
	assert.Equal(t, user.GivenName, accessClaims["given_name"])
	assert.Equal(t, user.GetFullName(), accessClaims["name"])
	assert.Equal(t, user.Username, accessClaims["preferred_username"])
	assert.Equal(t, fmt.Sprintf("%v/account/profile", config.Get().BaseURL), accessClaims["profile"])
	assert.Equal(t, "openid profile resource1:read authserver:userinfo", accessClaims["scope"])
	_, err = uuid.Parse(accessClaims["jti"].(string))
	assert.NoError(t, err)
	assertTimeClaimWithinRange(t, accessClaims, "updated_at", -1*time.Hour, "updated_at should be 1 hour ago")

	assertTimeClaimWithinRange(t, accessClaims, "iat", 0*time.Second, "iat should be now")
	assertTimeClaimWithinRange(t, accessClaims, "exp", 900*time.Second, "exp should be 900 seconds from now")
	assertTimeClaimWithinRange(t, accessClaims, "auth_time", -300*time.Second, "auth_time should be 300 seconds ago")

	// validate Refresh token --------------------------------------------

	refreshClaims := verifyAndDecodeToken(t, response.RefreshToken, publicKeyBytes)
	assert.Equal(t, user.Subject.String(), refreshClaims["sub"])
	assert.Equal(t, "https://test-issuer.com", refreshClaims["aud"])
	assert.Equal(t, "https://test-issuer.com", refreshClaims["iss"])
	assert.Equal(t, "Refresh", refreshClaims["typ"])
	assert.Equal(t, "openid profile resource1:read authserver:userinfo", refreshClaims["scope"])
	_, err = uuid.Parse(refreshClaims["jti"].(string))
	assert.NoError(t, err)
	assert.Equal(t, sessionIdentifier, refreshClaims["sid"])

	assertTimeClaimWithinRange(t, refreshClaims, "exp", 600*time.Second, "exp should be 600 seconds from now")
	assertTimeClaimWithinRange(t, refreshClaims, "iat", 0*time.Second, "iat should be now")

	// validate Refresh token passed to CreateRefreshToken --------------------------------------------

	assert.NotNil(t, capturedRefreshToken)
	assert.Equal(t, code.Id, capturedRefreshToken.CodeId)
	assert.NotEmpty(t, capturedRefreshToken.RefreshTokenJti)
	assert.Equal(t, refreshToken.FirstRefreshTokenJti, capturedRefreshToken.FirstRefreshTokenJti)
	assert.Equal(t, refreshToken.RefreshTokenJti, capturedRefreshToken.PreviousRefreshTokenJti)
	assert.Equal(t, "Refresh", capturedRefreshToken.RefreshTokenType)
	assert.Equal(t, "openid profile resource1:read authserver:userinfo", capturedRefreshToken.Scope)
	assert.Equal(t, sessionIdentifier, capturedRefreshToken.SessionIdentifier)
	assert.False(t, capturedRefreshToken.Revoked)
	assert.True(t, capturedRefreshToken.IssuedAt.Valid)
	assert.WithinDuration(t, now, capturedRefreshToken.IssuedAt.Time, 1*time.Second)
	assert.True(t, capturedRefreshToken.ExpiresAt.Valid)
	assert.WithinDuration(t, now.Add(600*time.Second), capturedRefreshToken.ExpiresAt.Time, 1*time.Second)

	mockDB.AssertExpectations(t)
}

func getTestPublicKey(t *testing.T) []byte {
	publicKeyBase64 := "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUNJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBZzhBTUlJQ0NnS0NBZ0VBb2Q3dFRUeVlCUjI0aDg1WEZaUkcKcUg4QXNuRTBXVHliRFZELzRYajdWSHY3RWErclU1S2cyVFdKNjhkL09BcCtNK1lMYnVMYXdIVk1mWWtQM0lhWgpTR1d6cHdCVXhxc0lmWTlLbEdaWjN3aGJIWi9EZWRCU2I1UjNkdnJjUEIvQmcrMkFucUVnRkV2N3Y4djZFT1psClJNU0RxR0t2VEU0a083UUFUR0JZbUJoTDZmNytPUnlRRmNLdUFZZ29DZmlKOG1hb3FkK1dIREk5TWMyTlBncncKMzhOaW1mQ3ZFM3VWbXljdFFyMDN3TTAyT1A0M0IyS3pCdUREc2ZKdUZWSVFWTUVtU2IyQ2ZqMjloWkpGMCtJWgpVZWYvVHhXQTZyVWpTK1pYSGtudXlYNDBVb1pYYUFJVU5zbUVEeVUxWHRKRTh5Ym1xSHdNK3BjT2dCSzh3TElXCk5LVkVFSDVtMjBsaStqMjdnSThvcVlNamJCYjdyYlI1L2JnSmdjL05qU2c4bTZrZDJzVC9TSmltMlI2eENFOEgKeVdTbEdvQkxIZTlSQUJYcUR4UTg5RCtiTGRRb1V5R1N2RVJVWXBBNzZYNGViY2tqdnR0UFl3cTFSWEZuS0VzbwpoUTQ0SVFySEFMTTRmcTQyRXF2WkZvTUxQVmhvT0xOSWd2NUlhU1lHZm9IMW1uQlZPZkJzZ3B4ejk0czRCNTJyCjFNcE5GaE1qaG1SSXBCcWYvSHNPalNtM05mUG1pYkVVQ0c2OEo3aSthU3ZvVTdwSnVZQzgyQW1TWmwxeWxLOTcKRjRUUG1RaGJUNG5yMlZxMS9oMGpwQUIzNW5DUS9tM09Sckl6RXYzL0F0UEdnbktlWENML3M0ZUQzd2hzbkNaTAowVmVMNmVFYWhoUFYydW05VlZzeVowVUNBd0VBQVE9PQotLS0tLUVORCBSU0EgUFVCTElDIEtFWS0tLS0tCg=="
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	assert.NoError(t, err)

	return publicKeyBytes
}

func getTestPrivateKey(t *testing.T) []byte {
	privateKeyBase64 := "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKSndJQkFBS0NBZ0VBb2Q3dFRUeVlCUjI0aDg1WEZaUkdxSDhBc25FMFdUeWJEVkQvNFhqN1ZIdjdFYStyClU1S2cyVFdKNjhkL09BcCtNK1lMYnVMYXdIVk1mWWtQM0lhWlNHV3pwd0JVeHFzSWZZOUtsR1paM3doYkhaL0QKZWRCU2I1UjNkdnJjUEIvQmcrMkFucUVnRkV2N3Y4djZFT1psUk1TRHFHS3ZURTRrTzdRQVRHQlltQmhMNmY3KwpPUnlRRmNLdUFZZ29DZmlKOG1hb3FkK1dIREk5TWMyTlBncnczOE5pbWZDdkUzdVZteWN0UXIwM3dNMDJPUDQzCkIyS3pCdUREc2ZKdUZWSVFWTUVtU2IyQ2ZqMjloWkpGMCtJWlVlZi9UeFdBNnJValMrWlhIa251eVg0MFVvWlgKYUFJVU5zbUVEeVUxWHRKRTh5Ym1xSHdNK3BjT2dCSzh3TElXTktWRUVINW0yMGxpK2oyN2dJOG9xWU1qYkJiNwpyYlI1L2JnSmdjL05qU2c4bTZrZDJzVC9TSmltMlI2eENFOEh5V1NsR29CTEhlOVJBQlhxRHhRODlEK2JMZFFvClV5R1N2RVJVWXBBNzZYNGViY2tqdnR0UFl3cTFSWEZuS0Vzb2hRNDRJUXJIQUxNNGZxNDJFcXZaRm9NTFBWaG8KT0xOSWd2NUlhU1lHZm9IMW1uQlZPZkJzZ3B4ejk0czRCNTJyMU1wTkZoTWpobVJJcEJxZi9Ic09qU20zTmZQbQppYkVVQ0c2OEo3aSthU3ZvVTdwSnVZQzgyQW1TWmwxeWxLOTdGNFRQbVFoYlQ0bnIyVnExL2gwanBBQjM1bkNRCi9tM09Sckl6RXYzL0F0UEdnbktlWENML3M0ZUQzd2hzbkNaTDBWZUw2ZUVhaGhQVjJ1bTlWVnN5WjBVQ0F3RUEKQVFLQ0FnQWZKS1hoWTFRWVArU2Q5RndhNGNGS2I4enhpQWc3VndhNTVDaW05OERiTzFOTnpzK1dyN0pVdUJGRwpGTWJzUUZDUnFhUHZmS1A3dlZXdkhXeTQwQWl6dmlWM2J2L2dqVTEvNHM3RmlIK29BcEtOTzR5L1pnNUdPM2xVCm9lVTNpQ0NTUW1LcG9uUnFrMGZuV2RaTjVCWDl5aFZPazFZSXgwdi9WSjF1RkdkWE0rMS9JcmxFd2JNVERMYXYKd3NONVQ2RXl5ditPVjE4cEk1MVVkS2pGRkJQTjZXaVNGNVdIbVJKcW5Ibi95aW5zNVU2V1hvcTEyQTU3dDBqUApkc1lwUWZXMGFNajJEUWtMUXRPdzNEaWxFRzR3clFNWTh4a3ZqeFF3YVN1L3Z4ZTdHcFgwZnJaWVkzWUNLSGxJCjlLNjFCSjJSYnAyWU11M0lWTUhNY0U1eWdKRDIxN0pLbzZKa3RRcTRoT09KRElhclhkSjIzT3N3OEdteEJyN1EKcSswQXhuVjd3NnRWRDVGOTg4SUpvc05OZW05cmgrUUN6YnhtN3BQc2JXT1hvdUlmQ2dyT0szZkJaZzg4QUs5UQpVRVZFSHlJUk5qMzBJSmw5MDh1d0JoWm9JVzJERk5xdERCQ1BJT09iZDFkblNLT2xlbjdRd3p0ajE1ak1QNm9oCnp3UU1pT0FHK051RmphY0FIRzRIWEx5NTZYK056RmJ6ZEZiWWZqTkZ5aElUd1Q5eC8vY3phSzVTcm8xY3ZFR0wKanJacFpXU1ZHOEJucVA4cGR3d3lwaml5KzM1alVGVnhkOWhVS3hBcjlHTkE3TFNsMG5qUHBRSTJiNnNSck92Tgp6KzlOR0h2UHI0N0F2WHJZazEvb01idHZLaHdmT0NqNjNZVDhOV1Z0YWxPRGtGNWdBUUtDQVFFQXlLc1RzUE4wCit4T0JYWDlDVW5kS25oemFIM1BaVnVxUmZrblJLbGZHVDFRYmQ2b1FLSWIwWWp6WUlDRWdTVGJNbTNMVWFOeEwKcUNqaktLbmRzL1ZWL0lSTjJYN1FaZHBtczFwU04zdnFSdnJxQzB4amp6RzhlZGhib1lKcGtZSHpoeENuc055eQpFNzZUTWpQSTdqTDNyQ0ROUDc3dkhJY0J1dUIrTUpyZVZXczN4RVFrR2ZaSldHTXNTNVRLY2pGQUpPV2JCaXhFCmY5WmJNNnRnc3lBUUdmU05sRXJXQll2c1kxemZtTVpWRmJIUTV1SmZkamUxN2l6TnNkNWI3c2dhbFR2L1ZXNW4KWXZKSHNaRzlKYldRdG9QTTQ3SFlmb21JMDhIMzV6K1M0YVZhdXVOWVpLT1RkYXB2ejRnOXE0dVhSS1plUFVlMgpLMHg0MXhUSzZQQ3ZSUUtDQVFFQXpvRXZoVldCdWZLeVVpZDBReExlSDVBQjNubkRNbndnUTJxTkh4VFdFS3pZCm5BVkVneDduZHY5OGorNjhqelZpNGhqY1Zxd2tvSTIxZWhKWktRVWdxaVo3ekg2bllrWE4wb2RVOVhUdHpzSGsKRjlkRWxsOFpSWkd4elFCMFFkN1dnb3pqcWF4YjN5YlZPOEN5WWFNbllsbTd3N1hxRllPNTVpLzBCL1c4UzQxVgpITEtSRXJya29ta2R0Q0pjNWhnYTg4eU5qNkV0SHhzL0J5aWcxRzhrc2RycGdmeVBWSTJMZnIrS3RwNGRKVmJOCmlLMnNQVmVMTU4wcXFRaTNSdCs0SjdMcFhLUllBM3A0M0FRLy9paXdmZ1NLOTEwVUVnbUlBVitjTGl6VU9mdkwKVk8xcnV3bmdlZlAvM2lwcWx0MWEyM3hnS1VIdmpZdlN3bUhFVlJwWUFRS0NBUUI3WHRLSVkrVnp4NVl0U1dRWgpGMFpFMXpBelRpSTlFWkhKdHRCbDIva01KSVdPbUh1K3J0bm8yOGQwV1dsa0dkREpjVnV0N0dLSFREdjhjQkxoCjVOK3NsQnJZc09LbS9CTlFDU09yQVFBVUM0ZUEwc0lTODEwUS9EZTVvRmdQSVhuN2UvM2MrcEp4R1NXZUk4QlEKMGZ6N1VsOWQ1YUZVUkp5SHJDVm85STNrcmpwbTdBM1YrRmszZ2lGbGhtREF2QTdYb0dJaTlXeFh2QTN1UWxyOQpSYVVnai8zTFFnYzYrYituaHgzZzYyNjhHOHAzYUkyUVBNZ1pXbXBNQkkwNHpNV3JJbXZrdGkvUjRXcTZmUU53Ci82T3Mwbk5ST2JJRWVjSXBqb00vSlJMRXI4aU1SZUcrWGVMMjRJWkZiVm1jOGdGYUwzNlk1bEhWWlBxV0lTNXUKOENxUkFvSUJBR3dBbkwzN2JwRzJJUlZlbFN2UFhtVGJpRjYzQ0NRTFQwUnpJY096dmhHU2xPZGt5ZVJaOFcwSApTanBzL2lsWUhwTnB0VE9QYk1pYjFPSTNYbkpad0MrOVdOb25FNXdPTGd1QnhDbHNNa1FFbkNycjUyOU41WVhCCklXQzZjQk5UWEpXQzRqOEhhalZYdGdZK1RnMUtxM3FBdS9jcjJYWFBJeGNFMVhpa1NRcXFyRzBKNTE0SWFUT1kKRG5UNzArSnprUVVaWXFCUUI2MVJMckdyeWhIUTN6dzE1aEtaNk15c0N0MExpSnppTFJRdVJlaktERjg0dmcrYwpYSWR6aTRlQjBtclE0OFFVSUFRUnRjdzhYTXVzdEVIMFZrbnhZR0hlb2tjMW5oVjRWTGJPdmhWNDV2TTN3ek9GCkxia2dMZ2NoVmplYzRSNHk0ZnNCdWdUMzVSc3RZQUVDZ2dFQUV5VTZkblRHZzFHZjM0Y0FRM1B0a09qMDRTblMKV0dVQTZPWTd4bHQ5WHBzclF0ekpNS0NaOEZiVWZBeVp2UkNvclhpN01BZzdwNTRUdWk5cmhvRlFMYW1ZRnJEZgpncEk3WjNCUVF3dkZCNUE3eWdQQmVRdzJHd2xkYVBuKzJrZTduZDZKdGIvZ1RJSjdFbmtLY041SXlmNHdqQ09TCjlWRmw3c2dldHMzMFFIYjlhZVBKTUJ4emlDM3N0K0x5azdEdmZMT2tOT0RvbHgyTE5aNW1hYkRiZ3BzLzlzdlIKU0ZYZEg5dGJRYWp5SEZnQnZCMVF6b3pSdTBnbGEvc0RHT0Z0MWtnQTE0OUlnM2ZSN3FhNGRIWDFoU012MmZPaQpNT1RFRDZxa1JwSHdGU3FsaVZPdzNPVkdOcnh5MGphWlhRSVZINUlqZmVFTUQwYnZzSS9uZ0lrTmFnPT0KLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K"
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	assert.NoError(t, err)

	return privateKeyBytes
}

func verifyAndDecodeToken(t *testing.T, tokenString string, publicKeyBytes []byte) jwt.MapClaims {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	}, jwt.WithExpirationRequired())
	assert.NoError(t, err)
	assert.True(t, token.Valid)
	return claims
}

func assertTimeClaimWithinRange(t *testing.T, claims jwt.MapClaims, claimName string, expectedDuration time.Duration, message string) {
	assert.Contains(t, claims, claimName)
	claimUnix := claims[claimName].(float64)
	claimTime := time.Unix(int64(claimUnix), 0)
	expectedTime := time.Now().UTC().Add(expectedDuration)
	start := expectedTime.Add(-3 * time.Second)
	end := expectedTime.Add(3 * time.Second)
	assert.True(t, claimTime.After(start) && claimTime.Before(end), fmt.Sprintf("%s: %s", message, claimTime))
}
