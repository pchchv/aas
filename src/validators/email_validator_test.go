package validators

import (
	"testing"

	"github.com/pchchv/aas/src/customerrors"
	"github.com/pchchv/aas/src/database/mocks"
	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	mockDB := mocks.NewDatabase(t)
	validator := NewEmailValidator(mockDB)
	tests := []struct {
		name          string
		email         string
		expectedError string
	}{
		{"Valid email", "test@example.com", ""},
		{"Invalid email - no @", "testexample.com", "Please enter a valid email address."},
		{"Invalid email - no domain", "test@.com", "Please enter a valid email address."},
		{"Invalid email - double dots", "test..email@example.com", "Please enter a valid email address."},
		{"Invalid email - starting with dot", ".test@example.com", "Please enter a valid email address."},
		{"Invalid email - ending with dot", "test.@example.com", "Please enter a valid email address."},
		{"Valid email with subdomains", "test@subdomain.example.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEmail(tt.email)
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
