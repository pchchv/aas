package enums

import "github.com/pkg/errors"

const (
	TokenTypeId TokenType = iota
	TokenTypeBearer
	TokenTypeRefresh

	AcrLevel1          AcrLevel = "urn:goiabada:level1"
	AcrLevel2Optional  AcrLevel = "urn:goiabada:level2_optional"
	AcrLevel2Mandatory AcrLevel = "urn:goiabada:level2_mandatory"

	PasswordPolicyNone   PasswordPolicy = iota // at least 1 char
	PasswordPolicyLow                          // at least 6 chars
	PasswordPolicyMedium                       // at least 8 chars. Must contain: 1 uppercase, 1 lowercase and 1 number
	PasswordPolicyHigh                         // at least 10 chars. Must contain: 1 uppercase, 1 lowercase, 1 number and 1 special character/symbol
)

type TokenType int

func (tt TokenType) String() string {
	return []string{"ID", "Bearer", "Refresh"}[tt]
}

type AcrLevel string

func (acrl AcrLevel) String() string {
	return string(acrl)
}

func AcrLevelFromString(s string) (AcrLevel, error) {
	switch s {
	case AcrLevel1.String():
		return AcrLevel1, nil
	case AcrLevel2Optional.String():
		return AcrLevel2Optional, nil
	case AcrLevel2Mandatory.String():
		return AcrLevel2Mandatory, nil
	}

	return "", errors.WithStack(errors.New("invalid ACR level " + s))
}

type PasswordPolicy int
