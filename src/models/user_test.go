package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_GetAddressClaim(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected map[string]string
	}{
		{"Empty address", User{}, map[string]string{}},
		{"Full address", User{
			AddressLine1:      "123 Main St",
			AddressLine2:      "Apt 4B",
			AddressLocality:   "Springfield",
			AddressRegion:     "IL",
			AddressPostalCode: "12345",
			AddressCountry:    "USA",
		}, map[string]string{
			"street_address": "123 Main St\r\nApt 4B",
			"locality":       "Springfield",
			"region":         "IL",
			"postal_code":    "12345",
			"country":        "USA",
			"formatted":      "123 Main St\r\nApt 4B\r\nSpringfield\r\nIL\r\n12345\r\nUSA",
		}},
		{"Partial address", User{
			AddressLine1:    "123 Main St",
			AddressLocality: "Springfield",
			AddressCountry:  "USA",
		}, map[string]string{
			"street_address": "123 Main St\r\n",
			"locality":       "Springfield",
			"country":        "USA",
			"formatted":      "123 Main St\r\n\r\nSpringfield\r\nUSA",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetAddressClaim()
			assert.Equal(t, tt.expected, result)

			// Additional check for the "formatted" field
			if formatted, ok := result["formatted"]; ok {
				assert.Equal(t, tt.expected["formatted"], formatted, "Formatted address mismatch")
			}
		})
	}
}
