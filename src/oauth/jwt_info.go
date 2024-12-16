package oauth

type JwtInfo struct {
	IdToken       *Jwt
	AccessToken   *Jwt
	RefreshToken  *Jwt
	TokenResponse TokenResponse
}
