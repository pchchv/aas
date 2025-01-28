package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateName(t *testing.T) {
	validator := NewProfileValidator(nil)
	tests := []struct {
		name      string
		inputName string
		nameField string
		wantErr   bool
	}{
		{"Valid name", "John Doe", "given name", false},
		{"Valid name with hyphen", "Mary-Jane", "given name", false},
		{"Valid name with apostrophe", "O'Connor", "family name", false},
		{"Too short", "A", "given name", true},
		{"Too long", "ThisNameIsTooLongAndExceedsTheMaximumAllowedLength", "given name", true},
		{"Invalid characters", "John123", "given name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateName(tt.inputName, tt.nameField)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
