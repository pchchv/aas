package validators

import (
	"regexp"

	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/database"
)

type ValidateProfileInput struct {
	Username            string
	GivenName           string
	MiddleName          string
	FamilyName          string
	Nickname            string
	Website             string
	Gender              string
	Locale              string
	DateOfBirth         string
	Subject             string
	ZoneInfo            string
	ZoneInfoCountryName string
}

type ProfileValidator struct {
	database database.Database
}

func NewProfileValidator(database database.Database) *ProfileValidator {
	return &ProfileValidator{
		database: database,
	}
}

func (val *ProfileValidator) ValidateName(name string, nameField string) error {
	pattern := `^[\p{L}\s'-]{2,48}$`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	if len(name) > 0 && !regex.MatchString(name) {
		return customerrors.NewErrorDetail(
			"",
			"Please enter a valid "+nameField+". It should contain only letters, spaces, hyphens, and apostrophes and be between 2 and 48 characters in length.",
		)
	}

	return nil
}
