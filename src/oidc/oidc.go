package oidc

func GetIdTokenScopeDescription(scope string) string {
	switch scope {
	case "openid":
		return "Authenticate your user and identify you via a unique ID"
	case "profile":
		return "Access to claims: name, family_name, given_name, middle_name, nickname, preferred_username, profile, website, gender, birthdate, zoneinfo, locale, and updated_at"
	case "email":
		return "Access to claims: email, email_verified"
	case "address":
		return "Access to the address claim"
	case "phone":
		return "Access to claims: phone_number and phone_number_verified"
	case "groups":
		return "Access to the list of groups that you belong to"
	case "attributes":
		return "Access to the attributes assigned to you by an admin, stored as key-value pairs"
	case "offline_access":
		return "Access to an offline refresh token, allowing the client to obtain a new access token without requiring your immediate interaction"
	default:
		return ""
	}
}
