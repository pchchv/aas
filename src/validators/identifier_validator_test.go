package validators

import (
	"testing"

	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/database/mocks"
	"github.com/stretchr/testify/assert"
)

func TestValidateIdentifier(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewIdentifierValidator(mockDB)
	tests := []struct {
		name             string
		identifier       string
		enforceMinLength bool
		expectedError    string
	}{
		{"Valid identifier", "valid-identifier123", true, ""},
		{"Valid identifier with underscore", "valid_identifier123", true, ""},
		{"Valid identifier minimum length", "abc", true, ""},
		{"Valid identifier not enforcing min length", "ab", false, ""},
		{"Too long identifier", "this-identifier-is-way-too-long-and-exceeds-maximum", true, "The identifier cannot exceed a maximum length of 38 characters."},
		{"Too short identifier", "ab", true, "The identifier must be at least 3 characters long."},
		{"Invalid start character", "1invalid-identifier", true, "Invalid identifier format. It must start with a letter, can include letters, numbers, dashes, and underscores, but cannot end with a dash or underscore, or have two consecutive dashes or underscores."},
		{"Invalid end character", "invalid-identifier-", true, "Invalid identifier format. It must start with a letter, can include letters, numbers, dashes, and underscores, but cannot end with a dash or underscore, or have two consecutive dashes or underscores."},
		{"Invalid end character underscore", "invalid_identifier_", true, "Invalid identifier format. It must start with a letter, can include letters, numbers, dashes, and underscores, but cannot end with a dash or underscore, or have two consecutive dashes or underscores."},
		{"Consecutive dashes", "invalid--identifier", true, "Invalid identifier format. It must start with a letter, can include letters, numbers, dashes, and underscores, but cannot end with a dash or underscore, or have two consecutive dashes or underscores."},
		{"Consecutive underscores", "invalid__identifier", true, "Invalid identifier format. It must start with a letter, can include letters, numbers, dashes, and underscores, but cannot end with a dash or underscore, or have two consecutive dashes or underscores."},
		{"Invalid characters", "invalid@identifier", true, "Invalid identifier format. It must start with a letter, can include letters, numbers, dashes, and underscores, but cannot end with a dash or underscore, or have two consecutive dashes or underscores."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateIdentifier(tt.identifier, tt.enforceMinLength)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				customErr, ok := err.(*customerrors.ErrorDetail)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError, customErr.GetDescription())
			}
		})
	}
}
