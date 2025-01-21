package oauth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pchchv/aas/src/database/mocks"
	"github.com/pchchv/aas/src/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDecodeAndValidateTokenString(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	tp := NewTokenParser(mockDB)
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey
	tests := []struct {
		name          string
		tokenClaims   jwt.MapClaims
		expectedError string
	}{
		{
			name: "Valid token",
			tokenClaims: jwt.MapClaims{
				"sub": "1234567890",
				"exp": time.Now().Add(time.Hour).Unix(),
			},
			expectedError: "",
		},
		{
			name: "Expired token",
			tokenClaims: jwt.MapClaims{
				"sub": "1234567890",
				"exp": time.Now().Add(-time.Hour).Unix(),
			},
			expectedError: "token has invalid claims: token is expired",
		},
		{
			name: "Missing expiration",
			tokenClaims: jwt.MapClaims{
				"sub": "1234567890",
			},
			expectedError: "token has invalid claims: token is missing required claim: exp claim is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, tt.tokenClaims)
			tokenString, _ := token.SignedString(privateKey)
			if result, err := tp.DecodeAndValidateTokenString(tokenString, publicKey, true); tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tokenString, result.TokenBase64)
				assert.Equal(t, tt.tokenClaims["sub"], result.Claims["sub"])
			}
		})
	}
}

func TestDecodeAndValidateTokenString_InvalidSignature(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	tp := NewTokenParser(mockDB)
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	wrongPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey
	claims := jwt.MapClaims{
		"sub": "1234567890",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString(wrongPrivateKey)
	result, err := tp.DecodeAndValidateTokenString(tokenString, publicKey, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token signature is invalid")
	assert.Nil(t, result)
}

func TestDecodeAndValidateTokenString_EmptyToken(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	tp := NewTokenParser(mockDB)
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey
	result, err := tp.DecodeAndValidateTokenString("", publicKey, true)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.TokenBase64)
	assert.Nil(t, result.Claims)
}

func TestDecodeAndValidateTokenResponse_EmptyTokens(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	tp := NewTokenParser(mockDB)
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKeyPEM := exportRSAPublicKeyAsPEMStr(&privateKey.PublicKey)
	mockDB.On("GetCurrentSigningKey", mock.Anything).Return(&models.KeyPair{
		PublicKeyPEM: []byte(publicKeyPEM),
	}, nil)

	tokenResponse := &TokenResponse{
		AccessToken:  "",
		IdToken:      "",
		RefreshToken: "",
	}

	token, err := tp.DecodeAndValidateTokenResponse(tokenResponse)
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Nil(t, token.AccessToken)
	assert.Nil(t, token.IdToken)
	assert.Nil(t, token.RefreshToken)
}

func createTestToken(privateKey *rsa.PrivateKey, claims map[string]interface{}, expirationTime time.Time) string {
	token := jwt.New(jwt.SigningMethodRS256)
	claims["exp"] = expirationTime.Unix()
	for k, v := range claims {
		token.Claims.(jwt.MapClaims)[k] = v
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}

	return tokenString
}

func exportRSAPublicKeyAsPEMStr(pubkey *rsa.PublicKey) string {
	pubkeyBytes, _ := x509.MarshalPKIXPublicKey(pubkey)
	pubkeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkeyBytes,
		},
	)
	return string(pubkeyPem)
}
