package validators

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/database"
)

type IdentifierValidator struct {
	database database.Database
}

func NewIdentifierValidator(database database.Database) *IdentifierValidator {
	return &IdentifierValidator{
		database: database,
	}
}

func (val *IdentifierValidator) ValidateIdentifier(identifier string, enforceMinLength bool) error {
	if maxLength := 38; len(identifier) > maxLength {
		return customerrors.NewErrorDetail("", fmt.Sprintf("The identifier cannot exceed a maximum length of %v characters.", maxLength))
	}

	if minLength := 3; enforceMinLength && len(identifier) < minLength {
		return customerrors.NewErrorDetail("", fmt.Sprintf("The identifier must be at least %v characters long.", minLength))
	}

	matchErrorMsg := "Invalid identifier format. It must start with a letter, can include letters, numbers, dashes, and underscores, but cannot end with a dash or underscore, or have two consecutive dashes or underscores."
	if match, err := regexp.MatchString("^[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$", identifier); err != nil {
		return err
	} else if !match {
		return customerrors.NewErrorDetail("", matchErrorMsg)
	}

	if strings.Contains(identifier, "--") || strings.Contains(identifier, "__") {
		return customerrors.NewErrorDetail("", matchErrorMsg)
	}

	return nil
}
