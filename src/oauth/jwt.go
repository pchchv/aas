package oauth

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Jwt struct {
	TokenBase64 string
	Claims      jwt.MapClaims
}

func (jwt Jwt) GetAddressClaim() map[string]string {
	if jwt.Claims["address"] != nil {
		if addressMap, ok := jwt.Claims["address"].(map[string]interface{}); ok {
			result := make(map[string]string)
			for k, v := range addressMap {
				result[k] = v.(string)
			}
			return result
		}
	}
	return map[string]string{}
}

func (jwt Jwt) GetStringClaim(claimName string) string {
	if jwt.Claims[claimName] != nil {
		return jwt.Claims[claimName].(string)
	}
	return ""
}

func (jwt Jwt) GetBoolClaim(claimName string) *bool {
	if jwt.Claims[claimName] != nil {
		if b, ok := jwt.Claims[claimName].(bool); ok {
			return &b
		}
	}
	return nil
}

func (jwt Jwt) GetTimeClaim(claimName string) time.Time {
	if jwt.Claims[claimName] != nil {
		if f64, ok := jwt.Claims[claimName].(float64); ok {
			return time.Unix(int64(f64), 0)
		}
	}

	return time.Unix(0, 0)
}

func (jwt Jwt) GetAudience() []string {
	if jwt.Claims["aud"] != nil {
		if audArr, ok := jwt.Claims["aud"].([]interface{}); ok {
			result := make([]string, len(audArr))
			for i, v := range audArr {
				result[i] = v.(string)
			}
			return result
		}

		if aud, ok := jwt.Claims["aud"].(string); ok {
			return []string{aud}
		}
	}

	return []string{}
}

func (jwt Jwt) HasScope(scope string) bool {
	if jwt.Claims["scope"] != nil {
		if scopesStr, ok := jwt.Claims["scope"].(string); ok {
			scopesArr := strings.Split(scopesStr, " ")
			for _, s := range scopesArr {
				if s == scope {
					return true
				}
			}
		}
	}
	return false
}

func (jwt Jwt) IsIssuerValid(issuer string) bool {
	return jwt.GetStringClaim("iss") == issuer
}
