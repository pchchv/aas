package validators

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/models"
	"github.com/pchchv/aas/src/oauth"
	"github.com/pchchv/aas/src/oidc"
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

type TokenValidator struct {
	database          database.Database
	tokenParser       TokenParser
	permissionChecker PermissionChecker
	auditLogger       AuditLogger
}

func NewTokenValidator(database database.Database, tokenParser TokenParser,
	permissionChecker PermissionChecker, auditLogger AuditLogger) *TokenValidator {
	return &TokenValidator{
		database:          database,
		tokenParser:       tokenParser,
		permissionChecker: permissionChecker,
		auditLogger:       auditLogger,
	}
}

func (val *TokenValidator) validateClientCredentialsScopes(scope string, client *models.Client) error {
	if len(scope) == 0 {
		return nil
	}

	space := regexp.MustCompile(`\s+`)
	scope = space.ReplaceAllString(scope, " ")
	scopes := strings.Split(scope, " ")
	for _, scopeStr := range scopes {
		if oidc.IsIdTokenScope(scopeStr) || oidc.IsOfflineAccessScope(scopeStr) {
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_request",
				fmt.Sprintf(
					"Id token scopes (such as '%v') are not supported in the client credentials flow. Please use scopes in the format 'resource:permission' (e.g., 'backendA:read'). Multiple scopes can be specified, separated by spaces.",
					scopeStr,
				),
				http.StatusBadRequest)
		}

		parts := strings.Split(scopeStr, ":")
		if len(parts) != 2 {
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_scope",
				fmt.Sprintf(
					"Invalid scope format: '%v'. Scopes must adhere to the resource-identifier:permission-identifier format. For instance: backend-service:create-product.",
					scopeStr,
				),
				http.StatusBadRequest)
		}

		res, err := val.database.GetResourceByResourceIdentifier(nil, parts[0])
		if err != nil {
			return err
		} else if res == nil {
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_scope",
				fmt.Sprintf("Invalid scope: '%v'. Could not find a resource with identifier '%v'.", scopeStr, parts[0]),
				http.StatusBadRequest)
		}

		permissions, err := val.database.GetPermissionsByResourceId(nil, res.Id)
		if err != nil {
			return err
		}

		permissionExists := false
		for _, perm := range permissions {
			if perm.PermissionIdentifier == parts[1] {
				permissionExists = true
				break
			}
		}

		if !permissionExists {
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_scope",
				fmt.Sprintf(
					"Scope '%v' is not recognized. The resource identified by '%v' doesn't grant the '%v' permission.",
					scopeStr,
					parts[0],
					parts[1],
				),
				http.StatusBadRequest)
		}

		clientHasPermission := false
		for _, perm := range client.Permissions {
			if perm.PermissionIdentifier == parts[1] {
				clientHasPermission = true
				break
			}
		}

		if !clientHasPermission {
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_scope",
				fmt.Sprintf("Permission to access scope '%v' is not granted to the client.", scopeStr),
				http.StatusBadRequest)
		}
	}

	return nil
}
