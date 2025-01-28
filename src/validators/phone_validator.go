package validators

import (
	"strings"

	"github.com/pchchv/aas/src/database"
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
