package validators

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
