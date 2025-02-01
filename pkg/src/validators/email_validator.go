package validators

import (
	"regexp"

	"github.com/pchchv/aas/pkg/src/customerrors"
	"github.com/pchchv/aas/pkg/src/database"
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

func (val *EmailValidator) ValidateEmailUpdate(input *ValidateEmailInput) error {
	if len(input.Email) == 0 {
		return customerrors.NewErrorDetail("", "Please enter an email address.")
	}

	if err := val.ValidateEmail(input.Email); err != nil {
		return err
	} else if len(input.Email) > 60 {
		return customerrors.NewErrorDetail("", "The email address cannot exceed a maximum length of 60 characters.")
	} else if input.Email != input.EmailConfirmation {
		return customerrors.NewErrorDetail("", "The email and email confirmation entries must be identical.")
	}

	user, err := val.database.GetUserBySubject(nil, input.Subject)
	if err != nil {
		return err
	}

	if userByEmail, err := val.database.GetUserByEmail(nil, input.Email); err != nil {
		return err
	} else if userByEmail != nil && userByEmail.Subject != user.Subject {
		return customerrors.NewErrorDetail("", "Apologies, but this email address is already registered.")
	}

	return nil
}
