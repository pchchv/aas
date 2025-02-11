package oauth

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pchchv/aas/pkg/src/config"
	"github.com/pchchv/aas/pkg/src/constants"
	"github.com/pchchv/aas/pkg/src/database"
	"github.com/pchchv/aas/pkg/src/enums"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pchchv/aas/pkg/src/oidc"
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

func (t *TokenIssuer) GenerateTokenResponseForAuthCode(ctx context.Context, code *models.Code) (*TokenResponse, error) {
	settings := ctx.Value(constants.ContextKeySettings).(*models.Settings)
	if err := t.database.CodeLoadClient(nil, code); err != nil {
		return nil, err
	}

	tokenExpirationInSeconds := settings.TokenExpirationInSeconds
	if code.Client.TokenExpirationInSeconds > 0 {
		tokenExpirationInSeconds = code.Client.TokenExpirationInSeconds
	}

	var tokenResponse = TokenResponse{
		TokenType: enums.TokenTypeBearer.String(),
		ExpiresIn: int64(tokenExpirationInSeconds),
	}

	keyPair, err := t.database.GetCurrentSigningKey(nil)
	if err != nil {
		return nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyPair.PrivateKeyPEM)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse private key from PEM")
	}

	now := time.Now().UTC()

	// access_token -----------------------------------------------------------------------

	if err = t.database.CodeLoadUser(nil, code); err != nil {
		return nil, err
	}

	if err = t.database.UserLoadGroups(nil, &code.User); err != nil {
		return nil, err
	}

	if err = t.database.GroupsLoadAttributes(nil, code.User.Groups); err != nil {
		return nil, err
	}

	if err = t.database.UserLoadAttributes(nil, &code.User); err != nil {
		return nil, err
	}

	accessTokenStr, scopeFromAccessToken, err := t.generateAccessToken(settings, code, code.Scope, now, privKey, keyPair.KeyIdentifier)
	if err != nil {
		return nil, err
	}
	tokenResponse.AccessToken = accessTokenStr
	tokenResponse.Scope = scopeFromAccessToken

	// id_token ---------------------------------------------------------------------------

	scopes := strings.Split(code.Scope, " ")
	if slices.Contains(scopes, "openid") {
		if idTokenStr, err := t.generateIdToken(settings, code, code.Scope, now, privKey, keyPair.KeyIdentifier); err != nil {
			return nil, err
		} else {
			tokenResponse.IdToken = idTokenStr
		}
	}

	// refresh_token ----------------------------------------------------------------------

	refreshToken, refreshExpiresIn, err := t.generateRefreshToken(settings, code, scopeFromAccessToken, now, privKey, keyPair.KeyIdentifier, nil)
	if err != nil {
		return nil, err
	}
	tokenResponse.RefreshToken = refreshToken
	tokenResponse.RefreshExpiresIn = refreshExpiresIn

	return &tokenResponse, nil
}

func (t *TokenIssuer) GenerateTokenResponseForClientCred(ctx context.Context, client *models.Client, scope string) (*TokenResponse, error) {
	settings := ctx.Value(constants.ContextKeySettings).(*models.Settings)
	var tokenResponse = TokenResponse{
		TokenType: "Bearer",
		ExpiresIn: int64(settings.TokenExpirationInSeconds),
		Scope:     scope,
	}

	keyPair, err := t.database.GetCurrentSigningKey(nil)
	if err != nil {
		return nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyPair.PrivateKeyPEM)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse private key from PEM")
	}

	now := time.Now().UTC()
	claims := make(jwt.MapClaims)
	scopes := strings.Split(scope, " ")

	// access_token ---------------------------------------------------------------------------

	claims["iss"] = settings.Issuer
	claims["sub"] = client.ClientIdentifier
	claims["iat"] = now.Unix()
	claims["jti"] = uuid.New().String()
	audCollection := []string{}
	for _, scope := range scopes {
		if oidc.IsIdTokenScope(scope) || oidc.IsOfflineAccessScope(scope) {
			continue
		}

		parts := strings.Split(scope, ":")
		if len(parts) != 2 {
			return nil, errors.WithStack(fmt.Errorf("invalid scope: %v", scope))
		}

		if !slices.Contains(audCollection, parts[0]) {
			audCollection = append(audCollection, parts[0])
		}
	}

	switch len(audCollection) {
	case 0:
		return nil, errors.WithStack(fmt.Errorf("unable to generate an access token without an audience. scope: '%v'", scope))
	case 1:
		claims["aud"] = audCollection[0]
	default:
		claims["aud"] = audCollection
	}

	claims["typ"] = enums.TokenTypeBearer.String()
	claims["exp"] = now.Add(time.Duration(time.Second * time.Duration(settings.TokenExpirationInSeconds))).Unix()
	claims["scope"] = scope
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyPair.KeyIdentifier
	accessToken, err := token.SignedString(privKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign access_token")
	}

	tokenResponse.AccessToken = accessToken
	return &tokenResponse, nil
}

func (t *TokenIssuer) GenerateTokenResponseForRefresh(ctx context.Context, input *GenerateTokenForRefreshInput) (*TokenResponse, error) {
	settings := ctx.Value(constants.ContextKeySettings).(*models.Settings)
	if err := t.database.CodeLoadClient(nil, input.Code); err != nil {
		return nil, err
	}

	scopeToUse := input.Code.Scope
	if len(input.ScopeRequested) > 0 {
		scopeToUse = input.ScopeRequested
	}

	tokenExpirationInSeconds := settings.TokenExpirationInSeconds
	if input.Code.Client.TokenExpirationInSeconds > 0 {
		tokenExpirationInSeconds = input.Code.Client.TokenExpirationInSeconds
	}

	var tokenResponse = TokenResponse{
		TokenType: enums.TokenTypeBearer.String(),
		ExpiresIn: int64(tokenExpirationInSeconds),
	}

	keyPair, err := t.database.GetCurrentSigningKey(nil)
	if err != nil {
		return nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyPair.PrivateKeyPEM)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse private key from PEM")
	}

	now := time.Now().UTC()

	// access_token -----------------------------------------------------------------------

	if err = t.database.CodeLoadUser(nil, input.Code); err != nil {
		return nil, err
	}

	if err = t.database.UserLoadGroups(nil, &input.Code.User); err != nil {
		return nil, err
	}

	if err = t.database.GroupsLoadAttributes(nil, input.Code.User.Groups); err != nil {
		return nil, err
	}

	if err = t.database.UserLoadAttributes(nil, &input.Code.User); err != nil {
		return nil, err
	}

	accessTokenStr, scopeFromAccessToken, err := t.generateAccessToken(settings, input.Code, scopeToUse, now, privKey, keyPair.KeyIdentifier)
	if err != nil {
		return nil, err
	}
	tokenResponse.AccessToken = accessTokenStr
	tokenResponse.Scope = scopeFromAccessToken

	// id_token ---------------------------------------------------------------------------

	scopes := strings.Split(scopeToUse, " ")
	if slices.Contains(scopes, "openid") {
		if idTokenStr, err := t.generateIdToken(settings, input.Code, scopeToUse, now, privKey, keyPair.KeyIdentifier); err != nil {
			return nil, err
		} else {
			tokenResponse.IdToken = idTokenStr
		}
	}

	// refresh_token ----------------------------------------------------------------------

	if refreshToken, refreshExpiresIn, err := t.generateRefreshToken(settings, input.Code, scopeFromAccessToken, now, privKey, keyPair.KeyIdentifier, input.RefreshToken); err != nil {
		return nil, err
	} else {
		tokenResponse.RefreshToken = refreshToken
		tokenResponse.RefreshExpiresIn = refreshExpiresIn
	}

	return &tokenResponse, nil
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

func (t *TokenIssuer) getRefreshTokenExpiration(refreshTokenType string, now time.Time, settings *models.Settings, client *models.Client) (int64, error) {
	if refreshTokenType == "Offline" {
		refreshTokenExpirationInSeconds := settings.RefreshTokenOfflineIdleTimeoutInSeconds
		if client.RefreshTokenOfflineIdleTimeoutInSeconds > 0 {
			refreshTokenExpirationInSeconds = client.RefreshTokenOfflineIdleTimeoutInSeconds
		}

		exp := now.Add(time.Duration(time.Second * time.Duration(refreshTokenExpirationInSeconds))).Unix()
		return exp, nil
	} else if refreshTokenType == "Refresh" {
		refreshTokenExpirationInSeconds := settings.UserSessionIdleTimeoutInSeconds
		exp := now.Add(time.Duration(time.Second * time.Duration(refreshTokenExpirationInSeconds))).Unix()
		return exp, nil
	}

	return 0, errors.WithStack(fmt.Errorf("invalid refresh token type: %v", refreshTokenType))
}

func (t *TokenIssuer) getRefreshTokenMaxLifetime(refreshTokenType string, now time.Time, settings *models.Settings, client *models.Client, sessionIdentifier string) (int64, error) {
	if refreshTokenType == "Offline" {
		maxLifetimeInSeconds := settings.RefreshTokenOfflineMaxLifetimeInSeconds
		if client.RefreshTokenOfflineMaxLifetimeInSeconds > 0 {
			maxLifetimeInSeconds = client.RefreshTokenOfflineMaxLifetimeInSeconds
		}

		maxLifetime := now.Add(time.Duration(time.Second * time.Duration(maxLifetimeInSeconds))).Unix()
		return maxLifetime, nil
	} else if refreshTokenType == "Refresh" {
		userSession, err := t.database.GetUserSessionBySessionIdentifier(nil, sessionIdentifier)
		if err != nil {
			return 0, err
		}

		maxLifetime := userSession.Started.Add(time.Duration(time.Second * time.Duration(settings.UserSessionMaxLifetimeInSeconds))).Unix()
		return maxLifetime, nil
	}

	return 0, errors.WithStack(fmt.Errorf("invalid refresh token type: %v", refreshTokenType))
}

func (t *TokenIssuer) generateRefreshToken(settings *models.Settings, code *models.Code, scope string, now time.Time, signingKey *rsa.PrivateKey, keyIdentifier string, refreshToken *models.RefreshToken) (string, int64, error) {
	claims := make(jwt.MapClaims)
	jti := uuid.New().String()
	claims["iss"] = settings.Issuer
	claims["iat"] = now.Unix()
	claims["jti"] = jti
	claims["aud"] = settings.Issuer
	claims["sub"] = code.User.Subject
	scopes := strings.Split(scope, " ")
	if slices.Contains(scopes, oidc.OfflineAccessScope) {
		// offline refresh token (not related to user session)
		claims["typ"] = "Offline"
		exp, err := t.getRefreshTokenExpiration("Offline", now, settings, &code.Client)
		if err != nil {
			return "", 0, err
		}

		maxLifetime, err := t.getRefreshTokenMaxLifetime("Offline", now, settings, &code.Client, code.SessionIdentifier)
		if err != nil {
			return "", 0, err
		}

		if refreshToken != nil {
			// if we are refreshing a refresh token, we need to use the max lifetime of the original refresh token
			maxLifetime = refreshToken.MaxLifetime.Time.Unix()
		}
		claims["offline_access_max_lifetime"] = maxLifetime

		if exp < maxLifetime {
			claims["exp"] = exp
		} else {
			claims["exp"] = maxLifetime
		}
	} else {
		// normal refresh token (associated with user session)
		claims["typ"] = "Refresh"
		claims["sid"] = code.SessionIdentifier
		exp, err := t.getRefreshTokenExpiration("Refresh", now, settings, &code.Client)
		if err != nil {
			return "", 0, err
		}

		maxLifetime, err := t.getRefreshTokenMaxLifetime("Refresh", now, settings, &code.Client, code.SessionIdentifier)
		if err != nil {
			return "", 0, err
		}

		if exp < maxLifetime {
			claims["exp"] = exp
		} else {
			claims["exp"] = maxLifetime
		}
	}
	claims["scope"] = scope

	// save 1st refresh token
	refreshTokenEntity := &models.RefreshToken{
		RefreshTokenJti:  jti,
		IssuedAt:         sql.NullTime{Time: now, Valid: true},
		ExpiresAt:        sql.NullTime{Time: time.Unix(claims["exp"].(int64), 0), Valid: true},
		CodeId:           code.Id,
		RefreshTokenType: claims["typ"].(string),
		Scope:            claims["scope"].(string),
		Revoked:          false,
	}
	if refreshToken != nil {
		refreshTokenEntity.PreviousRefreshTokenJti = refreshToken.RefreshTokenJti
		refreshTokenEntity.FirstRefreshTokenJti = refreshToken.FirstRefreshTokenJti
	} else {
		// first refresh token issued
		refreshTokenEntity.FirstRefreshTokenJti = jti
	}

	if slices.Contains(scopes, oidc.OfflineAccessScope) {
		t := time.Unix(claims["offline_access_max_lifetime"].(int64), 0)
		refreshTokenEntity.MaxLifetime = sql.NullTime{Time: t, Valid: true}
	} else {
		refreshTokenEntity.SessionIdentifier = claims["sid"].(string)
	}

	if err := t.database.CreateRefreshToken(nil, refreshTokenEntity); err != nil {
		return "", 0, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyIdentifier
	rt, err := token.SignedString(signingKey)
	if err != nil {
		return "", 0, errors.Wrap(err, "unable to sign refresh_token")
	}
	refreshExpiresIn := claims["exp"].(int64) - now.Unix()

	return rt, refreshExpiresIn, nil
}
