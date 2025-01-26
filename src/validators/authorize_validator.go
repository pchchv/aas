package validators

import "github.com/pchchv/aas/src/database"

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
