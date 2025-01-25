package user

import "github.com/pchchv/aas/src/database"

type CreateUserInput struct {
	Email         string
	EmailVerified bool
	GivenName     string
	MiddleName    string
	FamilyName    string
	PasswordHash  string
}

type UserCreator struct {
	database database.Database
}

func NewUserCreator(database database.Database) *UserCreator {
	return &UserCreator{
		database: database,
	}
}
