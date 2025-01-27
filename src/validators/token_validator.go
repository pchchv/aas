package validators

type PermissionChecker interface {
	UserHasScopePermission(userId int64, scope string) (bool, error)
}
