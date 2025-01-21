package oauth

import (
	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/models"
)

type GenerateTokenForRefreshInput struct {
	Code             *models.Code
	RefreshToken     *models.RefreshToken
	RefreshTokenInfo *Jwt
	ScopeRequested   string
}

type TokenIssuer struct {
	database    database.Database
	tokenParser *TokenParser
}

func NewTokenIssuer(database database.Database, tokenParser *TokenParser) *TokenIssuer {
	return &TokenIssuer{
		database:    database,
		tokenParser: tokenParser,
	}
}
