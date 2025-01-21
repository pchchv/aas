package oauth

import "github.com/pchchv/aas/src/models"

type GenerateTokenForRefreshInput struct {
	Code             *models.Code
	RefreshToken     *models.RefreshToken
	RefreshTokenInfo *Jwt
	ScopeRequested   string
}
