package oauth

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pchchv/aas/pkg/src/database"
	"github.com/pchchv/aas/pkg/src/hashutil"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pchchv/aas/pkg/src/stringutil"
)

type CreateCodeInput struct {
	AuthContext
	SessionIdentifier string
}

type CodeIssuer struct {
	database database.Database
}

func NewCodeIssuer(db database.Database) *CodeIssuer {
	return &CodeIssuer{
		database: db,
	}
}

func (ci *CodeIssuer) CreateAuthCode(input *CreateCodeInput) (*models.Code, error) {
	responseMode := input.ResponseMode
	if responseMode == "" {
		responseMode = "query"
	}

	client, err := ci.database.GetClientByClientIdentifier(nil, input.ClientId)
	if err != nil {
		return nil, err
	}

	var scope string
	space := regexp.MustCompile(`\s+`)
	if len(input.ConsentedScope) > 0 {
		scope = space.ReplaceAllString(input.ConsentedScope, " ")
	} else {
		scope = space.ReplaceAllString(input.Scope, " ")
	}

	scope = strings.TrimSpace(scope)
	authCode := strings.ReplaceAll(uuid.New().String(), "-", "") + stringutil.GenerateSecurityRandomString(96)
	authCodeHash, err := hashutil.HashString(authCode)
	if err != nil {
		return nil, err
	}

	code := &models.Code{
		Code:                authCode,
		CodeHash:            authCodeHash,
		ClientId:            client.Id,
		AuthenticatedAt:     time.Now().UTC(),
		UserId:              input.UserId,
		CodeChallenge:       input.CodeChallenge,
		CodeChallengeMethod: input.CodeChallengeMethod,
		RedirectURI:         input.RedirectURI,
		Scope:               scope,
		State:               input.State,
		Nonce:               input.Nonce,
		UserAgent:           input.UserAgent,
		ResponseMode:        responseMode,
		IpAddress:           input.IpAddress,
		AcrLevel:            input.AcrLevel,
		AuthMethods:         input.AuthMethods,
		SessionIdentifier:   input.SessionIdentifier,
		Used:                false,
	}

	if err = ci.database.CreateCode(nil, code); err != nil {
		return nil, err
	}

	return code, nil
}
