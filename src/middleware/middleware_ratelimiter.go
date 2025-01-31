package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/pchchv/aas/src/oauth"
)

type AuthHelper interface {
	GetAuthContext(r *http.Request) (*oauth.AuthContext, error)
}

type RateLimiterMiddleware struct {
	authHelper      AuthHelper
	pwdLimiter      *httprate.RateLimiter
	otpLimiter      *httprate.RateLimiter
	activateLimiter *httprate.RateLimiter
	resetPwdLimiter *httprate.RateLimiter
}

func NewRateLimiterMiddleware(authHelper AuthHelper) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		authHelper:      authHelper,
		pwdLimiter:      httprate.NewRateLimiter(10, 1*time.Minute),
		otpLimiter:      httprate.NewRateLimiter(10, 1*time.Minute),
		activateLimiter: httprate.NewRateLimiter(5, 5*time.Minute),
		resetPwdLimiter: httprate.NewRateLimiter(5, 5*time.Minute),
	}
}
