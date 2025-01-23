package oauth

import "github.com/pchchv/aas/src/database"

type CreateCodeInput struct {
	AuthContext
	SessionIdentifier string
}

type CodeIssuer struct {
	database database.Database
}

func NewCodeIssuer(db database.Database) *CodeIssuer {
	return &CodeIssuer{
		database: db,
	}
}
