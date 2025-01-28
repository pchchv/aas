package validators

import (
	"regexp"
	"strconv"
	"time"

	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/enums"
	"github.com/pchchv/aas/src/locales"
	"github.com/pchchv/aas/src/timezones"
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

func (val *ProfileValidator) ValidateProfile(input *ValidateProfileInput) (err error) {
	if len(input.Username) > 0 {
		user, err := val.database.GetUserBySubject(nil, input.Subject)
		if err != nil {
			return err
		}

		if userByUsername, err := val.database.GetUserByUsername(nil, input.Username); err != nil {
			return err
		} else if userByUsername != nil && userByUsername.Subject != user.Subject {
			return customerrors.NewErrorDetail("", "Sorry, this username is already taken.")
		}

		pattern := "^[a-zA-Z][a-zA-Z0-9_]{1,23}$"
		if regex, err := regexp.Compile(pattern); err != nil {
			return err
		} else if !regex.MatchString(input.Username) {
			return customerrors.NewErrorDetail(
				"",
				"Usernames must start with a letter and consist only of letters, numbers, and underscores. They must be between 2 and 24 characters long.",
			)
		}
	}

	if err = val.ValidateName(input.GivenName, "given name"); err != nil {
		return
	}

	if err = val.ValidateName(input.MiddleName, "middle name"); err != nil {
		return
	}

	if err = val.ValidateName(input.FamilyName, "family name"); err != nil {
		return
	}

	pattern := "^[a-zA-Z][a-zA-Z0-9_]{1,23}$"
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return
	}

	if len(input.Nickname) > 0 && !regex.MatchString(input.Nickname) {
		return customerrors.NewErrorDetail(
			"",
			"Nicknames must start with a letter and consist only of letters, numbers, and underscores. They must be between 2 and 24 characters long.",
		)
	}

	pattern = `^(https?://)?(www\.)?([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}(/\S*)?$`
	regex, err = regexp.Compile(pattern)
	if err != nil {
		return
	}

	if len(input.Website) > 0 && !regex.MatchString(input.Website) {
		return customerrors.NewErrorDetail("", "Please enter a valid website URL.")
	}

	if len(input.Website) > 96 {
		return customerrors.NewErrorDetail("", "Please ensure the website URL is no longer than 96 characters.")
	}

	if len(input.Gender) > 0 {
		if i, err := strconv.Atoi(input.Gender); err != nil {
			return customerrors.NewErrorDetail("", "Gender is invalid.")
		} else if !enums.IsGenderValid(i) {
			return customerrors.NewErrorDetail("", "Gender is invalid.")
		}
	}

	if len(input.DateOfBirth) > 0 {
		layout := "2006-01-02"
		if parsedTime, err := time.Parse(layout, input.DateOfBirth); err != nil {
			return customerrors.NewErrorDetail("", "The date of birth is invalid. Please use the format YYYY-MM-DD.")
		} else if parsedTime.After(time.Now()) {
			return customerrors.NewErrorDetail("", "The date of birth can't be in the future.")
		}
	}

	if len(input.ZoneInfo) > 0 {
		found := false
		timeZones := timezones.Get()
		for _, tz := range timeZones {
			if tz.Zone == input.ZoneInfo {
				found = true
				break
			}
		}

		if !found {
			return customerrors.NewErrorDetail("", "The zone info is invalid.")
		}
	}

	if len(input.Locale) > 0 {
		found := false
		locales := locales.Get()
		for _, loc := range locales {
			if loc.Id == input.Locale {
				found = true
				break
			}
		}

		if !found {
			return customerrors.NewErrorDetail("", "The locale is invalid.")
		}
	}

	return nil
}
