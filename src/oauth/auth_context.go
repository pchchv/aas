package oauth

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/pchchv/aas/src/enums"
	"github.com/pchchv/aas/src/models"
)

var (
	AuthStateInitial                 = "initial"
	AuthStateRequiresLevel1          = "requires_level_1"
	AuthStateRequiresLevel2          = "requires_level_2"
	AuthStateLevel1Password          = "level1_password"
	AuthStateLevel1PasswordCompleted = "level1_password_completed"
	AuthStateLevel1ExistingSession   = "level1_existing_session"
	AuthStateLevel2OTP               = "level2_otp"
	AuthStateLevel2OTPCompleted      = "level2_otp_completed"
	AuthStateAuthenticationCompleted = "authentication_completed"
	AuthStateRequiresConsent         = "requires_consent"
	AuthStateReadyToIssueCode        = "ready_to_issue_code"
)

type AuthContext struct {
	ClientId                      string
	RedirectURI                   string
	ResponseType                  string
	CodeChallengeMethod           string
	CodeChallenge                 string
	ResponseMode                  string
	Scope                         string
	ConsentedScope                string
	MaxAge                        string
	AcrValuesFromAuthorizeRequest string
	State                         string
	Nonce                         string
	UserAgent                     string
	IpAddress                     string
	AcrLevel                      string
	AuthMethods                   string
	AuthState                     string
	UserId                        int64
}

func (ac *AuthContext) HasScope(scope string) bool {
	if len(ac.Scope) == 0 {
		return false
	}
	return slices.Contains(strings.Split(ac.Scope, " "), scope)
}

func (ac *AuthContext) SetScope(scope string) {
	scopeArr := []string{}
	// remove duplicated spaces
	space := regexp.MustCompile(`\s+`)
	scopeSanitized := space.ReplaceAllString(scope, " ")
	// remove duplicated scopes
	scopeElements := strings.Split(scopeSanitized, " ")
	for _, s := range scopeElements {
		if !slices.Contains(scopeArr, strings.TrimSpace(s)) {
			scopeArr = append(scopeArr, strings.TrimSpace(s))
		}
	}

	ac.Scope = strings.TrimSpace(strings.Join(scopeArr, " "))
}

func (ac *AuthContext) SetAcrLevel(targetAcrLevel enums.AcrLevel, userSession *models.UserSession) (err error) {
	if userSession == nil {
		ac.AcrLevel = targetAcrLevel.String()
		return nil
	}

	userSessionAcrLevel, err := enums.AcrLevelFromString(userSession.AcrLevel)
	if err != nil {
		return
	}

	switch targetAcrLevel {
	case enums.AcrLevel1:
		if userSessionAcrLevel == enums.AcrLevel2Optional || userSessionAcrLevel == enums.AcrLevel2Mandatory {
			ac.AcrLevel = userSessionAcrLevel.String()
		} else {
			ac.AcrLevel = targetAcrLevel.String()
		}
	case enums.AcrLevel2Optional:
		if userSessionAcrLevel == enums.AcrLevel2Mandatory {
			ac.AcrLevel = userSessionAcrLevel.String()
		} else {
			ac.AcrLevel = targetAcrLevel.String()
		}
	default:
		ac.AcrLevel = targetAcrLevel.String()
	}

	return nil
}

func (ac *AuthContext) GetTargetAcrLevel(defaultAcrLevelFromClient enums.AcrLevel) enums.AcrLevel {
	acrValuesFromAuthorizeRequest := ac.parseAcrValuesFromAuthorizeRequest()
	if len(acrValuesFromAuthorizeRequest) > 0 {
		return acrValuesFromAuthorizeRequest[0]
	}
	return defaultAcrLevelFromClient
}

func (ac *AuthContext) ParseRequestedMaxAge() (requestedMaxAge *int) {
	if len(ac.MaxAge) > 0 {
		if i, err := strconv.Atoi(ac.MaxAge); err == nil {
			requestedMaxAge = &i
		}
	}

	return
}

func (ac *AuthContext) AddAuthMethod(method string) {
	if method = strings.ToLower(strings.TrimSpace(method)); method == "" {
		return
	}

	if ac.AuthMethods == "" {
		ac.AuthMethods = method
		return
	}

	lowerMethods := strings.ToLower(ac.AuthMethods)
	methods := strings.Fields(lowerMethods)
	for _, existingMethod := range methods {
		if existingMethod == method {
			return
		}
	}

	ac.AuthMethods = ac.AuthMethods + " " + method
}

func (ac *AuthContext) parseAcrValuesFromAuthorizeRequest() (arr []enums.AcrLevel) {
	acrValues := ac.AcrValuesFromAuthorizeRequest
	if len(strings.TrimSpace(acrValues)) > 0 {
		space := regexp.MustCompile(`\s+`)
		acrValues = space.ReplaceAllString(acrValues, " ")
		parts := strings.Split(acrValues, " ")
		for _, v := range parts {
			if acr, err := enums.AcrLevelFromString(v); err == nil && !slices.Contains(arr, acr) {
				arr = append(arr, acr)
			}
		}
	}

	return
}
