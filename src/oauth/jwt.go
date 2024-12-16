package oauth

import "github.com/golang-jwt/jwt/v5"

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
