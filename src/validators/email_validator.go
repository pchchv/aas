package validators

import "github.com/pchchv/aas/src/database"

type ValidateEmailInput struct {
	Email             string
	EmailConfirmation string
	Subject           string
}

type EmailValidator struct {
	database database.Database
}

func NewEmailValidator(database database.Database) *EmailValidator {
	return &EmailValidator{
		database: database,
	}
}
