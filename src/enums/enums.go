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

	KeyStateCurrent KeyState = iota
	KeyStatePrevious
	KeyStateNext

	ThreeStateSettingOn ThreeStateSetting = iota
	ThreeStateSettingOff
	ThreeStateSettingDefault

	AuthMethodPassword AuthMethod = iota
	AuthMethodOTP
)

type TokenType int

func (tt TokenType) String() string {
	return []string{"ID", "Bearer", "Refresh"}[tt]
}

type AcrLevel string

func (acrl AcrLevel) String() string {
	return string(acrl)
}

type PasswordPolicy int

func (p PasswordPolicy) String() string {
	return []string{"none", "low", "medium", "high"}[p]
}

type KeyState int

func (ks KeyState) String() string {
	return []string{"current", "previous", "next"}[ks]
}

type ThreeStateSetting int

func (tss ThreeStateSetting) String() string {
	return []string{"on", "off", "default"}[tss]
}

type AuthMethod int

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

func PasswordPolicyFromString(s string) (PasswordPolicy, error) {
	switch s {
	case PasswordPolicyNone.String():
		return PasswordPolicyNone, nil
	case PasswordPolicyLow.String():
		return PasswordPolicyLow, nil
	case PasswordPolicyMedium.String():
		return PasswordPolicyMedium, nil
	case PasswordPolicyHigh.String():
		return PasswordPolicyHigh, nil
	default:
		return PasswordPolicyNone, errors.WithStack(errors.New("invalid password policy " + s))
	}
}

func KeyStateFromString(s string) (KeyState, error) {
	switch s {
	case KeyStateCurrent.String():
		return KeyStateCurrent, nil
	case KeyStatePrevious.String():
		return KeyStatePrevious, nil
	case KeyStateNext.String():
		return KeyStateNext, nil
	default:
		return KeyStateCurrent, errors.WithStack(errors.New("invalid key state " + s))
	}
}

func ThreeStateSettingFromString(s string) (ThreeStateSetting, error) {
	switch s {
	case ThreeStateSettingOn.String():
		return ThreeStateSettingOn, nil
	case ThreeStateSettingOff.String():
		return ThreeStateSettingOff, nil
	case ThreeStateSettingDefault.String():
		return ThreeStateSettingDefault, nil
	default:
		return ThreeStateSettingOn, errors.WithStack(errors.New("invalid three state setting " + s))
	}
}
