package validators

import "github.com/pchchv/aas/src/database"

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
