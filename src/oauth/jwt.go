package oauth

import "github.com/golang-jwt/jwt/v5"

type Jwt struct {
	TokenBase64 string
	Claims      jwt.MapClaims
}
