package oauth

import (
	"testing"
	"time"

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

func TestGetBoolClaim(t *testing.T) {
	jwt := Jwt{Claims: map[string]interface{}{"bool": true}}
	assert.Equal(t, true, *jwt.GetBoolClaim("bool"))
	assert.Nil(t, jwt.GetBoolClaim("nonexistent"))
}

func TestGetTimeClaim(t *testing.T) {
	now := time.Now().Unix()
	jwt := Jwt{Claims: map[string]interface{}{"time": float64(now)}}
	assert.Equal(t, time.Unix(now, 0), jwt.GetTimeClaim("time"))
	assert.Equal(t, time.Time{}, jwt.GetTimeClaim("nonexistent"))
}

func TestGetAudience(t *testing.T) {
	tests := []struct {
		name     string
		claims   map[string]interface{}
		expected []string
	}{
		{
			name:     "No audience",
			claims:   map[string]interface{}{},
			expected: []string{},
		},
		{
			name:     "Single audience string",
			claims:   map[string]interface{}{"aud": "aud1"},
			expected: []string{"aud1"},
		},
		{
			name:     "Multiple audience array",
			claims:   map[string]interface{}{"aud": []interface{}{"aud1", "aud2"}},
			expected: []string{"aud1", "aud2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwt := Jwt{Claims: tt.claims}
			assert.Equal(t, tt.expected, jwt.GetAudience())
		})
	}
}

func TestHasScope(t *testing.T) {
	jwt := Jwt{Claims: map[string]interface{}{"scope": "read write"}}
	assert.True(t, jwt.HasScope("read"))
	assert.True(t, jwt.HasScope("write"))
	assert.False(t, jwt.HasScope("delete"))
	assert.False(t, Jwt{Claims: map[string]interface{}{}}.HasScope("read"))
}
