package oauth

import (
	"regexp"
	"slices"
	"strings"
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
