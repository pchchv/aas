package validators

type ValidatePhoneInput struct {
	PhoneNumber          string
	PhoneNumberVerified  bool
	PhoneCountryUniqueId string
}
