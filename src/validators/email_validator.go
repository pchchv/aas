package validators

import (
	"regexp"

	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/database"
)

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

func (val *EmailValidator) ValidateEmail(emailAddress string) error {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` // basic regex pattern for email validation
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	if !regex.MatchString(emailAddress) {
		return customerrors.NewErrorDetail("", "Please enter a valid email address.")
	}

	atIndex := regexp.MustCompile("@").FindStringIndex(emailAddress)
	localPart := emailAddress[:atIndex[0]]

	if regexp.MustCompile(`\.\.`).MatchString(emailAddress) {
		return customerrors.NewErrorDetail("", "Please enter a valid email address.")
	}

	if localPart[0] == '.' || localPart[len(localPart)-1] == '.' {
		return customerrors.NewErrorDetail("", "Please enter a valid email address.")
	}

	return nil
}
