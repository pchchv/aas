package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAddressClaim(t *testing.T) {
	address := map[string]interface{}{"street": "123 Main St", "city": "Anytown"}
	jwt := Jwt{Claims: map[string]interface{}{"address": address}}
	expected := map[string]string{"street": "123 Main St", "city": "Anytown"}
	assert.Equal(t, expected, jwt.GetAddressClaim())
	assert.Empty(t, Jwt{Claims: map[string]interface{}{}}.GetAddressClaim())
}

func TestGetStringClaim(t *testing.T) {
	jwt := Jwt{Claims: map[string]interface{}{"test": "value"}}
	assert.Equal(t, "value", jwt.GetStringClaim("test"))
	assert.Equal(t, "", jwt.GetStringClaim("nonexistent"))
}
