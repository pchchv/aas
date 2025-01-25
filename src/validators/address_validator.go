package validators

import (
	"github.com/biter777/countries"
	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/database"
)

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

func (val *AddressValidator) ValidateAddress(input *ValidateAddressInput) error {
	if len(input.AddressLine1) > 60 {
		return customerrors.NewErrorDetail("", "Please ensure the address line 1 is no longer than 60 characters.")
	}

	if len(input.AddressLine2) > 60 {
		return customerrors.NewErrorDetail("", "Please ensure the address line 2 is no longer than 60 characters.")
	}

	if len(input.AddressLocality) > 60 {
		return customerrors.NewErrorDetail("", "Please ensure the locality is no longer than 60 characters.")
	}

	if len(input.AddressRegion) > 60 {
		return customerrors.NewErrorDetail("", "Please ensure the region is no longer than 60 characters.")
	}

	if len(input.AddressPostalCode) > 30 {
		errorMsg := "Please ensure the postal code is no longer than 30 characters."
		return customerrors.NewErrorDetail("", errorMsg)
	}

	if len(input.AddressCountry) > 0 {
		if country := countries.ByName(input.AddressCountry); country.Info().Code == 0 {
			return customerrors.NewErrorDetail("", "Invalid country.")
		}
	}

	return nil
}
