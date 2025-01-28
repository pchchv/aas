package validators

import "github.com/pchchv/aas/src/database"

type ValidatePhoneInput struct {
	PhoneNumber          string
	PhoneNumberVerified  bool
	PhoneCountryUniqueId string
}

type PhoneValidator struct {
	database database.Database
}

func NewPhoneValidator(database database.Database) *PhoneValidator {
	return &PhoneValidator{
		database: database,
	}
}
