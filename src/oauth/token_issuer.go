package oauth

import (
	"fmt"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pchchv/aas/src/config"
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

func (t *TokenIssuer) addClaimIfNotEmpty(claims jwt.MapClaims, claimName string, claimValue string) {
	if len(strings.TrimSpace(claimValue)) > 0 {
		claims[claimName] = claimValue
	}
}

func (t *TokenIssuer) addOpenIdConnectClaims(claims jwt.MapClaims, code *models.Code) {
	scopes := strings.Split(code.Scope, " ")
	if len(scopes) > 1 || (len(scopes) == 1 && scopes[0] != "openid") {
		claims["updated_at"] = code.User.UpdatedAt.Time.UTC().Unix()
	}

	if slices.Contains(scopes, "profile") {
		t.addClaimIfNotEmpty(claims, "name", code.User.GetFullName())
		t.addClaimIfNotEmpty(claims, "given_name", code.User.GivenName)
		t.addClaimIfNotEmpty(claims, "middle_name", code.User.MiddleName)
		t.addClaimIfNotEmpty(claims, "family_name", code.User.FamilyName)
		t.addClaimIfNotEmpty(claims, "nickname", code.User.Nickname)
		t.addClaimIfNotEmpty(claims, "preferred_username", code.User.Username)
		claims["profile"] = fmt.Sprintf("%v/account/profile", config.Get().BaseURL)
		t.addClaimIfNotEmpty(claims, "website", code.User.Website)
		t.addClaimIfNotEmpty(claims, "gender", code.User.Gender)
		if code.User.BirthDate.Valid {
			claims["birthdate"] = code.User.BirthDate.Time.Format("2006-01-02")
		}
		t.addClaimIfNotEmpty(claims, "zoneinfo", code.User.ZoneInfo)
		t.addClaimIfNotEmpty(claims, "locale", code.User.Locale)
	}

	if slices.Contains(scopes, "email") {
		t.addClaimIfNotEmpty(claims, "email", code.User.Email)
		claims["email_verified"] = code.User.EmailVerified
	}

	if slices.Contains(scopes, "address") && code.User.HasAddress() {
		claims["address"] = code.User.GetAddressClaim()
	}

	if slices.Contains(scopes, "phone") {
		t.addClaimIfNotEmpty(claims, "phone_number", code.User.PhoneNumber)
		claims["phone_number_verified"] = code.User.PhoneNumberVerified
	}
}
