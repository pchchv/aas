package user

import "github.com/pchchv/aas/src/database"

type PermissionChecker struct {
	db database.Database
}

func NewPermissionChecker(database database.Database) *PermissionChecker {
	return &PermissionChecker{
		db: database,
	}
}
