package user

import (
	"github.com/google/uuid"
	"github.com/pchchv/aas/pkg/src/constants"
	"github.com/pchchv/aas/pkg/src/database"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

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

func (uc *UserCreator) CreateUser(input *CreateUserInput) (*models.User, error) {
	user := &models.User{
		Subject:       uuid.New(),
		Enabled:       true,
		Email:         input.Email,
		EmailVerified: input.EmailVerified,
		GivenName:     input.GivenName,
		MiddleName:    input.MiddleName,
		FamilyName:    input.FamilyName,
		PasswordHash:  input.PasswordHash,
	}

	authServerResource, err := uc.database.GetResourceByResourceIdentifier(nil, constants.AdminConsoleResourceIdentifier)
	if err != nil {
		return nil, err
	}

	permissions, err := uc.database.GetPermissionsByResourceId(nil, authServerResource.Id)
	if err != nil {
		return nil, err
	}

	var accountPermission *models.Permission
	for idx, permission := range permissions {
		if permission.PermissionIdentifier == constants.ManageAccountPermissionIdentifier {
			accountPermission = &permissions[idx]
			break
		}
	}

	if accountPermission == nil {
		return nil, errors.WithStack(errors.New("unable to find the account permission"))
	}

	user.Permissions = []models.Permission{*accountPermission}
	tx, err := uc.database.BeginTransaction()
	if err != nil {
		return nil, err
	}
	defer uc.database.RollbackTransaction(tx) //nolint:errcheck

	if err = uc.database.CreateUser(tx, user); err != nil {
		return nil, err
	}

	for _, permission := range user.Permissions {
		up := &models.UserPermission{
			UserId:       user.Id,
			PermissionId: permission.Id,
		}
		if err = uc.database.CreateUserPermission(tx, up); err != nil {
			return nil, err
		}
	}

	if err = uc.database.CommitTransaction(tx); err != nil {
		return nil, err
	}

	return user, nil
}
