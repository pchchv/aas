package validators

import "github.com/pchchv/aas/src/database"

type IdentifierValidator struct {
	database database.Database
}

func NewIdentifierValidator(database database.Database) *IdentifierValidator {
	return &IdentifierValidator{
		database: database,
	}
}
