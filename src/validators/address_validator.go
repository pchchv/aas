package validators

import "github.com/pchchv/aas/src/database"

type ValidateAddressInput struct {
	AddressLine1      string
	AddressLine2      string
	AddressLocality   string
	AddressRegion     string
	AddressPostalCode string
	AddressCountry    string
}

type AddressValidator struct {
	database database.Database
}

func NewAddressValidator(database database.Database) *AddressValidator {
	return &AddressValidator{
		database: database,
	}
}
