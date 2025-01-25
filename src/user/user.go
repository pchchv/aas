package user

type CreateUserInput struct {
	Email         string
	EmailVerified bool
	GivenName     string
	MiddleName    string
	FamilyName    string
	PasswordHash  string
}
