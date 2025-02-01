package validators

import (
	"regexp"
	"strings"

	"github.com/pchchv/aas/pkg/src/customerrors"
	"github.com/pchchv/aas/pkg/src/database"
	"github.com/pchchv/aas/pkg/src/phones"
)

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

func (val *PhoneValidator) ValidatePhone(input *ValidatePhoneInput) error {
	if len(input.PhoneCountryUniqueId) > 0 {
		pc := phones.Get()
		found := false
		for _, c := range pc {
			if c.UniqueId == input.PhoneCountryUniqueId {
				found = true
				break
			}
		}

		if !found {
			return customerrors.NewErrorDetail("", "Phone country is invalid.")
		}

		if len(input.PhoneNumber) == 0 {
			return customerrors.NewErrorDetail("", "The phone number field must contain a valid phone number. To remove the phone number information, please select the (blank) option from the dropdown menu for the phone country and leave the phone number field empty.")
		}
	}

	if len(input.PhoneNumber) > 0 {
		// Remove spaces and hyphens for length check and pattern matching
		cleanNumber := strings.ReplaceAll(strings.ReplaceAll(input.PhoneNumber, " ", ""), "-", "")

		// Check minimum length
		if len(cleanNumber) < 6 {
			return customerrors.NewErrorDetail("", "The phone number must be at least 6 digits long.")
		}

		// Check for simple patterns
		if isSimplePattern(cleanNumber) {
			return customerrors.NewErrorDetail("", "The phone number appears to be a simple pattern. Please enter a valid phone number.")
		}

		pattern := `^[0-9]+([- ]?[0-9]+)*$`
		if regex, err := regexp.Compile(pattern); err != nil {
			return err
		} else if !regex.MatchString(input.PhoneNumber) {
			return customerrors.NewErrorDetail("", "Please enter a valid number. Phone numbers can contain only digits, and may include single spaces or hyphens as separators.")
		} else if len(input.PhoneNumber) > 30 {
			return customerrors.NewErrorDetail("", "The maximum allowed length for a phone number is 30 characters.")
		} else if len(input.PhoneCountryUniqueId) == 0 {
			return customerrors.NewErrorDetail("", "You must select a country for your phone number.")
		}
	}

	return nil
}

func isSimplePattern(number string) bool {
	// Check for all repeated digits (e.g., 00000, 111111111, etc.)
	if len(number) > 0 && strings.Count(number, string(number[0])) == len(number) {
		return true
	}

	// Check for sequential ascending digits
	ascending := "0123456789"
	if strings.Contains(ascending, number) {
		return true
	}

	// Check for sequential descending digits
	descending := "9876543210"
	return strings.Contains(descending, number)
}
