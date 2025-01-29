package handlerhelpers

import "github.com/gorilla/sessions"

type AuthHelper struct {
	sessionStore sessions.Store
}

func NewAuthHelper(sessionStore sessions.Store) *AuthHelper {
	return &AuthHelper{
		sessionStore: sessionStore,
	}
}
