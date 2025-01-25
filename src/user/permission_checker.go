package user

import (
	"strings"

	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

type PermissionChecker struct {
	db database.Database
}

func NewPermissionChecker(database database.Database) *PermissionChecker {
	return &PermissionChecker{
		db: database,
	}
}

func (pc *PermissionChecker) UserHasScopePermission(userId int64, scope string) (bool, error) {
	user, err := pc.db.GetUserById(nil, userId)
	if err != nil {
		return false, err
	} else if user == nil {
		return false, nil
	}

	if err = pc.db.UserLoadPermissions(nil, user); err != nil {
		return false, err
	}

	if err = pc.db.UserLoadGroups(nil, user); err != nil {
		return false, err
	}

	if err = pc.db.GroupsLoadPermissions(nil, user.Groups); err != nil {
		return false, err
	}

	parts := strings.Split(scope, ":")
	if len(parts) != 2 {
		return false, errors.WithStack(errors.New("invalid scope format: " + scope + ". expected format: resource_identifier:permission_identifier"))
	}

	resourceIdentifier := parts[0]
	permissionIdentifier := parts[1]
	resource, err := pc.db.GetResourceByResourceIdentifier(nil, resourceIdentifier)
	if err != nil {
		return false, err
	} else if resource == nil {
		return false, err
	}

	permissions, err := pc.db.GetPermissionsByResourceId(nil, resource.Id)
	if err != nil {
		return false, err
	}

	var mp *models.Permission
	for idx, p := range permissions {
		if p.PermissionIdentifier == permissionIdentifier {
			mp = &permissions[idx]
			break

		}
	}

	if mp == nil {
		return false, err
	}

	userHasPermission := false
	for _, userPerm := range user.Permissions {
		if userPerm.Id == mp.Id {
			userHasPermission = true
			break
		}
	}

	if userHasPermission {
		return true, nil
	}

	groupHasPermission := false
	for _, group := range user.Groups {
		for _, groupPerm := range group.Permissions {
			if groupPerm.Id == mp.Id {
				groupHasPermission = true
				break
			}
		}
	}

	return groupHasPermission, nil
}
