package middleware

import (
	"net/http"

	"github.com/pchchv/aas/src/oauth"
)

type AuthHelper interface {
	GetAuthContext(r *http.Request) (*oauth.AuthContext, error)
}
