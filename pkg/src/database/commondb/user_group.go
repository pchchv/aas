package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateUserGroup(tx *sql.Tx, userGroup *models.UserGroup) error {
	if userGroup.UserId == 0 {
		return errors.WithStack(errors.New("can't create userGroup with user_id 0"))
	}

	if userGroup.GroupId == 0 {
		return errors.WithStack(errors.New("can't create userGroup with group_id 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := userGroup.CreatedAt
	originalUpdatedAt := userGroup.UpdatedAt
	userGroup.CreatedAt = sql.NullTime{Time: now, Valid: true}
	userGroup.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	userGroupStruct := sqlbuilder.NewStruct(new(models.UserGroup)).For(d.Flavor)
	insertBuilder := userGroupStruct.WithoutTag("pk").InsertInto("users_groups", userGroup)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		userGroup.CreatedAt = originalCreatedAt
		userGroup.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert userGroup")
	}

	id, err := result.LastInsertId()
	if err != nil {
		userGroup.CreatedAt = originalCreatedAt
		userGroup.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	userGroup.Id = id
	return nil
}

func (d *CommonDB) GetUserGroupsByUserIds(tx *sql.Tx, userIds []int64) (userGroups []models.UserGroup, err error) {
	if len(userIds) == 0 {
		return nil, nil
	}

	userGroupStruct := sqlbuilder.NewStruct(new(models.UserGroup)).For(d.Flavor)
	selectBuilder := userGroupStruct.SelectFrom("users_groups")
	selectBuilder.Where(selectBuilder.In("user_id", sqlbuilder.Flatten(userIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userGroup models.UserGroup
		addr := userGroupStruct.Addr(&userGroup)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userGroup")
		}
		userGroups = append(userGroups, userGroup)
	}

	return
}

func (d *CommonDB) GetUserGroupsByUserId(tx *sql.Tx, userId int64) (userGroups []models.UserGroup, err error) {
	userGroupStruct := sqlbuilder.NewStruct(new(models.UserGroup)).For(d.Flavor)
	selectBuilder := userGroupStruct.SelectFrom("users_groups")
	selectBuilder.Where(selectBuilder.Equal("user_id", userId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userGroup models.UserGroup
		addr := userGroupStruct.Addr(&userGroup)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userGroup")
		}
		userGroups = append(userGroups, userGroup)
	}

	return
}

func (d *CommonDB) GetUserGroupByUserIdAndGroupId(tx *sql.Tx, userId, groupId int64) (*models.UserGroup, error) {
	userGroupStruct := sqlbuilder.NewStruct(new(models.UserGroup)).For(d.Flavor)
	selectBuilder := userGroupStruct.SelectFrom("users_groups")
	selectBuilder.Where(selectBuilder.Equal("user_id", userId))
	selectBuilder.Where(selectBuilder.Equal("group_id", groupId))
	return d.getUserGroupCommon(tx, selectBuilder, userGroupStruct)
}

func (d *CommonDB) GetUserGroupById(tx *sql.Tx, userGroupId int64) (*models.UserGroup, error) {
	userGroupStruct := sqlbuilder.NewStruct(new(models.UserGroup)).For(d.Flavor)
	selectBuilder := userGroupStruct.SelectFrom("users_groups")
	selectBuilder.Where(selectBuilder.Equal("id", userGroupId))
	return d.getUserGroupCommon(tx, selectBuilder, userGroupStruct)
}

func (d *CommonDB) UpdateUserGroup(tx *sql.Tx, userGroup *models.UserGroup) error {
	if userGroup.Id == 0 {
		return errors.WithStack(errors.New("can't update userGroup with id 0"))
	}

	originalUpdatedAt := userGroup.UpdatedAt
	userGroup.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	userGroupStruct := sqlbuilder.NewStruct(new(models.UserGroup)).For(d.Flavor)
	updateBuilder := userGroupStruct.WithoutTag("pk").WithoutTag("dont-update").Update("users_groups", userGroup)
	updateBuilder.Where(updateBuilder.Equal("id", userGroup.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		userGroup.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update userGroup")
	}

	return nil
}

func (d *CommonDB) DeleteUserGroup(tx *sql.Tx, userGroupId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.UserGroup)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("users_groups")
	deleteBuilder.Where(deleteBuilder.Equal("id", userGroupId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete userGroup")
	}

	return nil
}

func (d *CommonDB) getUserGroupCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, userGroupStruct *sqlbuilder.Struct) (*models.UserGroup, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var userGroup models.UserGroup
	if rows.Next() {
		addr := userGroupStruct.Addr(&userGroup)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userGroup")
		}
		return &userGroup, nil
	}

	return nil, nil
}
