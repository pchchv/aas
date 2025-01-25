package user

import (
	"strings"

	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/models"
	"github.com/pchchv/aas/src/oidc"
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

func (pc *PermissionChecker) FilterOutScopesWhereUserIsNotAuthorized(scope string, user *models.User) (string, error) {
	if user == nil {
		return "", errors.WithStack(errors.New("user is nil"))
	}

	var newScope string
	for _, scopeStr := range strings.Split(scope, " ") {
		if scopeStr != "" {
			if oidc.IsIdTokenScope(scopeStr) || oidc.IsOfflineAccessScope(scopeStr) {
				newScope += scopeStr + " "
			} else {
				parts := strings.Split(scopeStr, ":")
				if len(parts) != 2 {
					return "", errors.WithStack(errors.New("invalid scope format: " + scopeStr))
				} else {
					if userHasPermission, err := pc.UserHasScopePermission(user.Id, scopeStr); err != nil {
						return "", err
					} else if userHasPermission {
						newScope += scopeStr + " "
					}
				}
			}
		}
	}

	return strings.TrimSpace(newScope), nil
}
