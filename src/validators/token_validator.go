package validators

import (
	"crypto/rsa"

	"github.com/pchchv/aas/src/models"
	"github.com/pchchv/aas/src/oauth"
)

type PermissionChecker interface {
	UserHasScopePermission(userId int64, scope string) (bool, error)
}

type TokenParser interface {
	DecodeAndValidateTokenString(token string, pubKey *rsa.PublicKey, withExpirationCheck bool) (*oauth.Jwt, error)
}

type AuditLogger interface {
	Log(auditEvent string, details map[string]interface{})
}

type ValidateTokenRequestInput struct {
	GrantType    string
	Code         string
	RedirectURI  string
	CodeVerifier string
	ClientId     string
	ClientSecret string
	Scope        string
	RefreshToken string
}

type ValidateTokenRequestResult struct {
	Scope            string
	Client           *models.Client
	CodeEntity       *models.Code
	RefreshToken     *models.RefreshToken
	RefreshTokenInfo *oauth.Jwt
}
