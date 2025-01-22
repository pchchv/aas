package oauth

import (
	"crypto/rsa"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pchchv/aas/src/config"
	"github.com/pchchv/aas/src/constants"
	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/enums"
	"github.com/pchchv/aas/src/models"
	"github.com/pchchv/aas/src/oidc"
	"github.com/pkg/errors"
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

func (t *TokenIssuer) generateIdToken(settings *models.Settings, code *models.Code, scope string,
	now time.Time, signingKey *rsa.PrivateKey, keyIdentifier string) (idToken string, err error) {
	claims := make(jwt.MapClaims)
	claims["iss"] = settings.Issuer
	claims["sub"] = code.User.Subject
	claims["iat"] = now.Unix()
	claims["auth_time"] = code.AuthenticatedAt.Unix()
	claims["jti"] = uuid.New().String()
	claims["acr"] = code.AcrLevel
	claims["amr"] = code.AuthMethods
	claims["sid"] = code.SessionIdentifier
	scopes := strings.Split(scope, " ")
	claims["aud"] = code.Client.ClientIdentifier
	claims["typ"] = enums.TokenTypeId.String()
	tokenExpirationInSeconds := settings.TokenExpirationInSeconds
	if code.Client.TokenExpirationInSeconds > 0 {
		tokenExpirationInSeconds = code.Client.TokenExpirationInSeconds
	}

	claims["exp"] = now.Add(time.Duration(time.Second * time.Duration(tokenExpirationInSeconds))).Unix()
	if len(code.Nonce) > 0 {
		claims["nonce"] = code.Nonce
	}
	t.addOpenIdConnectClaims(claims, code)

	// groups
	if slices.Contains(scopes, "groups") {
		groups := []string{}
		for _, group := range code.User.Groups {
			if group.IncludeInIdToken {
				groups = append(groups, group.GroupIdentifier)
			}
		}

		if len(groups) > 0 {
			claims["groups"] = groups
		}
	}

	// attributes
	if slices.Contains(scopes, "attributes") {
		attributes := map[string]string{}
		for _, attribute := range code.User.Attributes {
			if attribute.IncludeInIdToken {
				attributes[attribute.Key] = attribute.Value
			}
		}

		for _, group := range code.User.Groups {
			for _, attribute := range group.Attributes {
				if attribute.IncludeInIdToken {
					attributes[attribute.Key] = attribute.Value
				}
			}
		}

		if len(attributes) > 0 {
			claims["attributes"] = attributes
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyIdentifier
	if idToken, err = token.SignedString(signingKey); err != nil {
		return "", errors.Wrap(err, "unable to sign id_token")
	}

	return
}

func (t *TokenIssuer) generateAccessToken(settings *models.Settings, code *models.Code, scope string,
	now time.Time, signingKey *rsa.PrivateKey, keyIdentifier string) (string, string, error) {
	claims := make(jwt.MapClaims)
	claims["iss"] = settings.Issuer
	claims["sub"] = code.User.Subject
	claims["iat"] = now.Unix()
	claims["auth_time"] = code.AuthenticatedAt.Unix()
	claims["jti"] = uuid.New().String()
	claims["acr"] = code.AcrLevel
	claims["amr"] = code.AuthMethods
	claims["sid"] = code.SessionIdentifier
	scopes := strings.Split(scope, " ")
	addUserInfoScope := false
	audCollection := []string{}
	for _, s := range scopes {
		if oidc.IsIdTokenScope(s) {
			// if an OIDC scope is present, give access to the userinfo endpoint
			if !slices.Contains(audCollection, constants.AuthServerResourceIdentifier) {
				audCollection = append(audCollection, constants.AuthServerResourceIdentifier)
			}
			addUserInfoScope = true
			continue
		}

		if !oidc.IsOfflineAccessScope(s) {
			parts := strings.Split(s, ":")
			if len(parts) != 2 {
				return "", "", errors.WithStack(fmt.Errorf("invalid scope: %v", s))
			}
			if !slices.Contains(audCollection, parts[0]) {
				audCollection = append(audCollection, parts[0])
			}
		}
	}

	switch len(audCollection) {
	case 0:
		return "", "", errors.WithStack(fmt.Errorf("unable to generate an access token without an audience. scope: '%v'", scope))
	case 1:
		claims["aud"] = audCollection[0]
	default:
		claims["aud"] = audCollection
	}

	if addUserInfoScope {
		// if an OIDC scope is present, give access to the userinfo endpoint
		userInfoScopeStr := fmt.Sprintf("%v:%v", constants.AuthServerResourceIdentifier, constants.UserinfoPermissionIdentifier)
		if !slices.Contains(scopes, userInfoScopeStr) {
			scopes = append(scopes, userInfoScopeStr)
		}
		scope = strings.Join(scopes, " ")
	}

	claims["typ"] = enums.TokenTypeBearer.String()
	tokenExpirationInSeconds := settings.TokenExpirationInSeconds
	if code.Client.TokenExpirationInSeconds > 0 {
		tokenExpirationInSeconds = code.Client.TokenExpirationInSeconds
	}

	claims["exp"] = now.Add(time.Duration(time.Second * time.Duration(tokenExpirationInSeconds))).Unix()
	claims["scope"] = scope
	if len(code.Nonce) > 0 {
		claims["nonce"] = code.Nonce
	}

	includeOpenIDConnectClaimsInAccessToken := settings.IncludeOpenIDConnectClaimsInAccessToken
	if code.Client.IncludeOpenIDConnectClaimsInAccessToken == enums.ThreeStateSettingOn.String() ||
		code.Client.IncludeOpenIDConnectClaimsInAccessToken == enums.ThreeStateSettingOff.String() {
		includeOpenIDConnectClaimsInAccessToken = code.Client.IncludeOpenIDConnectClaimsInAccessToken == enums.ThreeStateSettingOn.String()
	}

	if slices.Contains(scopes, "openid") && includeOpenIDConnectClaimsInAccessToken {
		t.addOpenIdConnectClaims(claims, code)
	}

	// groups
	if slices.Contains(scopes, "groups") {
		groups := []string{}
		for _, group := range code.User.Groups {
			if group.IncludeInAccessToken {
				groups = append(groups, group.GroupIdentifier)
			}
		}

		if len(groups) > 0 {
			claims["groups"] = groups
		}
	}

	// attributes
	if slices.Contains(scopes, "attributes") {
		attributes := map[string]string{}
		for _, attribute := range code.User.Attributes {
			if attribute.IncludeInAccessToken {
				attributes[attribute.Key] = attribute.Value
			}
		}

		for _, group := range code.User.Groups {
			for _, attribute := range group.Attributes {
				if attribute.IncludeInAccessToken {
					attributes[attribute.Key] = attribute.Value
				}
			}
		}

		if len(attributes) > 0 {
			claims["attributes"] = attributes
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyIdentifier
	accessToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to sign access_token")
	}

	return accessToken, scope, nil
}
