package user

import (
	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/oauth"
)

type UserSessionManager struct {
	codeIssuer   *oauth.CodeIssuer
	sessionStore sessions.Store
	database     database.Database
}

func NewUserSessionManager(codeIssuer *oauth.CodeIssuer, sessionStore sessions.Store, database database.Database) *UserSessionManager {
	return &UserSessionManager{
		codeIssuer:   codeIssuer,
		sessionStore: sessionStore,
		database:     database,
	}
}
