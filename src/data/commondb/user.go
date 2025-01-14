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
