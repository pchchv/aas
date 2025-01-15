package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateUser(tx *sql.Tx, user *models.User) error {
	now := time.Now().UTC()
	originalCreatedAt := user.CreatedAt
	originalUpdatedAt := user.UpdatedAt
	user.CreatedAt = sql.NullTime{Time: now, Valid: true}
	user.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	insertBuilder := userStruct.WithoutTag("pk").InsertInto("users", user)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		user.CreatedAt = originalCreatedAt
		user.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert user")
	}

	id, err := result.LastInsertId()
	if err != nil {
		user.CreatedAt = originalCreatedAt
		user.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	user.Id = id
	return nil
}

func (d *CommonDB) GetUserById(tx *sql.Tx, userId int64) (*models.User, error) {
	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	selectBuilder := userStruct.SelectFrom("users")
	selectBuilder.Where(selectBuilder.Equal("id", userId))
	return d.getUserCommon(tx, selectBuilder, userStruct)
}

func (d *CommonDB) GetUsersByIds(tx *sql.Tx, userIds []int64) (users map[int64]models.User, err error) {
	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	selectBuilder := userStruct.SelectFrom("users")
	selectBuilder.Where(selectBuilder.In("id", sqlbuilder.Flatten(userIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		addr := userStruct.Addr(&user)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan user")
		}
		users[user.Id] = user
	}

	return
}

func (d *CommonDB) UserLoadAttributes(tx *sql.Tx, user *models.User) error {
	if user != nil {
		if userAttributes, err := d.GetUserAttributesByUserId(tx, user.Id); err != nil {
			return err
		} else {
			user.Attributes = userAttributes
		}
	}
	return nil
}

func (d *CommonDB) UsersLoadPermissions(tx *sql.Tx, users []models.User) error {
	if users == nil {
		return nil
	}

	userIds := make([]int64, len(users))
	for i, user := range users {
		userIds[i] = user.Id
	}

	userPermissions, err := d.GetUserPermissionsByUserIds(tx, userIds)
	if err != nil {
		return err
	}

	permissionIds := make([]int64, len(userPermissions))
	for i, userPermission := range userPermissions {
		permissionIds[i] = userPermission.PermissionId
	}

	permissions, err := d.GetPermissionsByIds(tx, permissionIds)
	if err != nil {
		return err
	}

	permissionMap := make(map[int64]models.Permission)
	for _, permission := range permissions {
		permissionMap[permission.Id] = permission
	}

	permissionsByUserId := make(map[int64][]models.Permission)
	for _, userPermission := range userPermissions {
		if permission, ok := permissionMap[userPermission.PermissionId]; ok {
			permissionsByUserId[userPermission.UserId] = append(permissionsByUserId[userPermission.UserId], permission)
		}
	}

	for i, user := range users {
		users[i].Permissions = permissionsByUserId[user.Id]
	}

	return nil
}

func (d *CommonDB) UserLoadPermissions(tx *sql.Tx, user *models.User) error {
	if user == nil {
		return nil
	}

	userPermissions, err := d.GetUserPermissionsByUserId(tx, user.Id)
	if err != nil {
		return err
	}

	permissionIds := make([]int64, len(userPermissions))
	for i, userPermission := range userPermissions {
		permissionIds[i] = userPermission.PermissionId
	}

	permissions, err := d.GetPermissionsByIds(tx, permissionIds)
	if err != nil {
		return err
	}

	user.Permissions = permissions
	return nil
}

func (d *CommonDB) GetUserByUsername(tx *sql.Tx, username string) (*models.User, error) {
	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	selectBuilder := userStruct.SelectFrom("users")
	selectBuilder.Where(selectBuilder.Equal("username", username))
	return d.getUserCommon(tx, selectBuilder, userStruct)
}

func (d *CommonDB) GetUserBySubject(tx *sql.Tx, subject string) (*models.User, error) {
	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	selectBuilder := userStruct.SelectFrom("users")
	selectBuilder.Where(selectBuilder.Equal("subject", subject))
	return d.getUserCommon(tx, selectBuilder, userStruct)
}

func (d *CommonDB) GetUserByEmail(tx *sql.Tx, email string) (*models.User, error) {
	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	selectBuilder := userStruct.SelectFrom("users")
	selectBuilder.Where(selectBuilder.Equal("email", email))
	return d.getUserCommon(tx, selectBuilder, userStruct)
}

func (d *CommonDB) GetLastUserWithOTPState(tx *sql.Tx, otpEnabledState bool) (*models.User, error) {
	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	selectBuilder := userStruct.SelectFrom("users")
	selectBuilder.Where(
		selectBuilder.And(
			selectBuilder.Equal("otp_enabled", otpEnabledState),
			selectBuilder.Equal("enabled", true),
		),
	)
	selectBuilder.OrderBy("id").Desc()
	selectBuilder.Limit(1)
	return d.getUserCommon(tx, selectBuilder, userStruct)
}

func (d *CommonDB) UpdateUser(tx *sql.Tx, user *models.User) error {
	if user.Id == 0 {
		return errors.WithStack(errors.New("can't update user with id 0"))
	}

	originalUpdatedAt := user.UpdatedAt
	user.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	updateBuilder := userStruct.WithoutTag("pk").WithoutTag("dont-update").Update("users", user)
	updateBuilder.Where(updateBuilder.Equal("id", user.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		user.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update user")
	}

	return nil
}

func (d *CommonDB) SearchUsersPaginated(tx *sql.Tx, query string, page int, pageSize int) (users []models.User, count int, err error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	selectBuilder := userStruct.SelectFrom("users")
	if query != "" {
		selectBuilder.Where(
			selectBuilder.Or(
				selectBuilder.Like("subject", "%"+query+"%"),
				selectBuilder.Like("username", "%"+query+"%"),
				selectBuilder.Like("given_name", "%"+query+"%"),
				selectBuilder.Like("middle_name", "%"+query+"%"),
				selectBuilder.Like("family_name", "%"+query+"%"),
				selectBuilder.Like("email", "%"+query+"%"),
			),
		)
	}
	selectBuilder.OrderBy("users.given_name").Asc()
	selectBuilder.Offset((page - 1) * pageSize)
	selectBuilder.Limit(pageSize)
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		addr := userStruct.Addr(&user)
		if err = rows.Scan(addr...); err != nil {
			return nil, 0, errors.Wrap(err, "unable to scan user")
		}
		users = append(users, user)
	}

	selectBuilder = d.Flavor.NewSelectBuilder()
	selectBuilder.Select("count(*)").From("users")
	if query != "" {
		selectBuilder.Where(
			selectBuilder.Or(
				selectBuilder.Like("subject", "%"+query+"%"),
				selectBuilder.Like("username", "%"+query+"%"),
				selectBuilder.Like("given_name", "%"+query+"%"),
				selectBuilder.Like("middle_name", "%"+query+"%"),
				selectBuilder.Like("family_name", "%"+query+"%"),
				selectBuilder.Like("email", "%"+query+"%"),
			),
		)
	}

	sql, args = selectBuilder.Build()
	rows2, err := d.QuerySql(nil, sql, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to query database")
	}
	defer rows2.Close()

	if rows2.Next() {
		err = rows2.Scan(&count)
		if err != nil {
			return nil, 0, errors.Wrap(err, "unable to scan count")
		}
	}

	return
}

func (d *CommonDB) DeleteUser(tx *sql.Tx, userId int64) error {
	userStruct := sqlbuilder.NewStruct(new(models.UserSession)).For(d.Flavor)
	deleteBuilder := userStruct.DeleteFrom("users")
	deleteBuilder.Where(deleteBuilder.Equal("id", userId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete user")
	}

	return nil
}

func (d *CommonDB) getUserCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, userStruct *sqlbuilder.Struct) (*models.User, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var user models.User
	if rows.Next() {
		addr := userStruct.Addr(&user)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan user")
		}
		return &user, nil
	}

	return nil, nil
}
