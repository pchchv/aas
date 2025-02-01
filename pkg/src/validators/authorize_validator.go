package validators

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/pchchv/aas/pkg/src/constants"
	"github.com/pchchv/aas/pkg/src/customerrors"
	"github.com/pchchv/aas/pkg/src/database"
	"github.com/pchchv/aas/pkg/src/oidc"
)

type ValidateClientAndRedirectURIInput struct {
	RequestId   string
	ClientId    string
	RedirectURI string
}

type ValidateRequestInput struct {
	ResponseType        string
	ResponseMode        string
	CodeChallenge       string
	CodeChallengeMethod string
}

type AuthorizeValidator struct {
	database database.Database
}

func NewAuthorizeValidator(database database.Database) *AuthorizeValidator {
	return &AuthorizeValidator{
		database: database,
	}
}

func (val *AuthorizeValidator) ValidateScopes(scope string) (err error) {
	scope = strings.TrimSpace(scope)
	if len(scope) == 0 {
		return customerrors.NewErrorDetailWithHttpStatusCode("invalid_scope",
			"The 'scope' parameter is missing. Ensure to include one or more scopes, separated by spaces. Scopes can be an OpenID Connect scope, a resource:permission scope, or a combination of both.",
			http.StatusBadRequest)
	}

	space := regexp.MustCompile(`\s+`)
	scope = space.ReplaceAllString(scope, " ")
	scopes := strings.Split(scope, " ")
	for _, scopeStr := range scopes {
		// these scopes don't need further validation
		if oidc.IsIdTokenScope(scopeStr) || oidc.IsOfflineAccessScope(scopeStr) {
			continue
		}

		userInfoScope := fmt.Sprintf("%v:%v", constants.AuthServerResourceIdentifier, constants.UserinfoPermissionIdentifier)
		if scopeStr == userInfoScope {
			err = errors.New("The '" + userInfoScope + "' scope is automatically included in the access token when an OpenID Connect scope is present. There's no need to request it explicitly. Please remove it from your request.")
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_scope", err.Error(), http.StatusBadRequest)
		}

		parts := strings.Split(scopeStr, ":")
		if len(parts) != 2 {
			err = errors.New("Invalid scope format: '" + scopeStr + "'. Scopes must adhere to the resource-identifier:permission-identifier format. For instance: backend-service:create-product.")
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_scope", err.Error(), http.StatusBadRequest)
		}

		res, err := val.database.GetResourceByResourceIdentifier(nil, parts[0])
		if err != nil {
			return err
		} else if res == nil {
			err = errors.New("Invalid scope: '" + scopeStr + "'. Could not find a resource with identifier '" + parts[0] + "'.")
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_scope", err.Error(), http.StatusBadRequest)
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
			err = errors.New("Scope '" + scopeStr + "' is invalid. The resource identified by '" + parts[0] + "' does not have a permission with identifier '" + parts[1] + "'.")
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_scope", err.Error(), http.StatusBadRequest)
		}
	}

	return nil
}

func (val *AuthorizeValidator) ValidateRequest(input *ValidateRequestInput) error {
	if input.ResponseType != "code" {
		return customerrors.NewErrorDetailWithHttpStatusCode("invalid_request", "Ensure response_type is set to 'code' as it's the only supported value.", http.StatusBadRequest)
	}

	if input.CodeChallengeMethod != "S256" {
		return customerrors.NewErrorDetailWithHttpStatusCode("invalid_request", "PKCE is required. Ensure code_challenge_method is set to 'S256'.", http.StatusBadRequest)
	}

	if len(input.CodeChallenge) < 43 || len(input.CodeChallenge) > 128 {
		return customerrors.NewErrorDetailWithHttpStatusCode("invalid_request", "The code_challenge parameter is either missing or incorrect. It should be 43 to 128 characters long.", http.StatusBadRequest)
	}

	if len(input.ResponseMode) > 0 {
		if !slices.Contains([]string{"query", "fragment", "form_post"}, input.ResponseMode) {
			return customerrors.NewErrorDetailWithHttpStatusCode("invalid_request", "Invalid response_mode parameter. Supported values are: query, fragment, form_post.", http.StatusBadRequest)
		}
	}

	return nil
}

func (val *AuthorizeValidator) ValidateClientAndRedirectURI(input *ValidateClientAndRedirectURIInput) (err error) {
	if len(input.ClientId) == 0 {
		return customerrors.NewErrorDetail("", "The client_id parameter is missing.")
	}

	client, err := val.database.GetClientByClientIdentifier(nil, input.ClientId)
	if err == nil {
		if client == nil {
			err = customerrors.NewErrorDetail("", "Invalid client_id parameter. The client does not exist.")
		} else if !client.Enabled {
			err = customerrors.NewErrorDetail("", "Invalid client_id parameter. The client is disabled.")
		} else if !client.AuthorizationCodeEnabled {
			err = customerrors.NewErrorDetail("", "Invalid client_id parameter. The client does not support the authorization code flow.")
		} else if len(input.RedirectURI) == 0 {
			err = customerrors.NewErrorDetail("", "The redirect_uri parameter is missing.")
		}
	}

	if err != nil {
		return
	}

	if err = val.database.ClientLoadRedirectURIs(nil, client); err != nil {
		return
	}

	clientHasRedirectURI := false
	for _, r := range client.RedirectURIs {
		if input.RedirectURI == r.URI {
			clientHasRedirectURI = true
		}
	}

	if !clientHasRedirectURI {
		return customerrors.NewErrorDetail("", "Invalid redirect_uri parameter. The client does not have this redirect URI registered.")
	}

	return nil
}
