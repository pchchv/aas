package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateGroup(tx *sql.Tx, group *models.Group) error {
	now := time.Now().UTC()
	originalCreatedAt := group.CreatedAt
	originalUpdatedAt := group.UpdatedAt
	group.CreatedAt = sql.NullTime{Time: now, Valid: true}
	group.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	groupStruct := sqlbuilder.NewStruct(new(models.Group)).For(d.Flavor)
	insertBuilder := groupStruct.WithoutTag("pk").InsertInto(d.Flavor.Quote("groups"), group)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		group.CreatedAt = originalCreatedAt
		group.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert group")
	}

	id, err := result.LastInsertId()
	if err != nil {
		group.CreatedAt = originalCreatedAt
		group.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	group.Id = id
	return nil
}

func (d *CommonDB) GetGroupsByIds(tx *sql.Tx, groupIds []int64) (groups []models.Group, err error) {
	if len(groupIds) == 0 {
		return nil, nil
	}

	groupStruct := sqlbuilder.NewStruct(new(models.Group)).For(d.Flavor)
	selectBuilder := groupStruct.SelectFrom(d.Flavor.Quote("groups"))
	selectBuilder.Where(selectBuilder.In("id", sqlbuilder.Flatten(groupIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var group models.Group
		addr := groupStruct.Addr(&group)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan group")
		}
		groups = append(groups, group)
	}

	return groups, nil
}

func (d *CommonDB) GetGroupById(tx *sql.Tx, groupId int64) (*models.Group, error) {
	groupStruct := sqlbuilder.NewStruct(new(models.Group)).For(d.Flavor)
	selectBuilder := groupStruct.SelectFrom(d.Flavor.Quote("groups"))
	selectBuilder.Where(selectBuilder.Equal("id", groupId))
	return d.getGroupCommon(tx, selectBuilder, groupStruct)
}

func (d *CommonDB) GetAllGroups(tx *sql.Tx) (groups []models.Group, err error) {
	groupStruct := sqlbuilder.NewStruct(new(models.Group)).For(d.Flavor)
	selectBuilder := groupStruct.SelectFrom(d.Flavor.Quote("groups"))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var group models.Group
		addr := groupStruct.Addr(&group)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan group")
		}

		groups = append(groups, group)
	}

	return
}

func (d *CommonDB) GetAllGroupsPaginated(tx *sql.Tx, page int, pageSize int) (groups []models.Group, total int, err error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	groupStruct := sqlbuilder.NewStruct(new(models.Group)).For(d.Flavor)
	selectBuilder := groupStruct.SelectFrom(d.Flavor.Quote("groups"))
	selectBuilder.OrderBy("group_identifier").Asc()
	selectBuilder.Offset((page - 1) * pageSize)
	selectBuilder.Limit(pageSize)
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var group models.Group
		addr := groupStruct.Addr(&group)
		if err = rows.Scan(addr...); err != nil {
			return nil, 0, errors.Wrap(err, "unable to scan group")
		}

		groups = append(groups, group)
	}

	selectBuilder = d.Flavor.NewSelectBuilder()
	selectBuilder.Select("count(*)").From(d.Flavor.Quote("groups"))
	sql, args = selectBuilder.Build()
	rows2, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to query database")
	}
	defer rows2.Close()

	if rows2.Next() {
		if err = rows2.Scan(&total); err != nil {
			return nil, 0, errors.Wrap(err, "unable to scan count")
		}
	}

	return
}

func (d *CommonDB) GetGroupMembersPaginated(tx *sql.Tx, groupId int64, page int, pageSize int) (users []models.User, total int, err error) {
	if groupId <= 0 {
		return nil, 0, errors.WithStack(errors.New("group id must be greater than 0"))
	}

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	userStruct := sqlbuilder.NewStruct(new(models.User)).For(d.Flavor)
	selectBuilder := userStruct.SelectFrom("users")
	selectBuilder.JoinWithOption(sqlbuilder.InnerJoin, "users_groups", "users.id = users_groups.user_id")
	selectBuilder.Where(selectBuilder.Equal("users_groups.group_id", groupId))
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
	selectBuilder.JoinWithOption(sqlbuilder.InnerJoin, "users_groups", "users.id = users_groups.user_id")
	selectBuilder.Where(selectBuilder.Equal("users_groups.group_id", groupId))
	sql, args = selectBuilder.Build()
	rows2, err := d.QuerySql(nil, sql, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to query database")
	}
	defer rows2.Close()

	if rows2.Next() {
		if err = rows2.Scan(&total); err != nil {
			return nil, 0, errors.Wrap(err, "unable to scan count")
		}
	}

	return
}

func (d *CommonDB) GetGroupByGroupIdentifier(tx *sql.Tx, groupIdentifier string) (*models.Group, error) {
	groupStruct := sqlbuilder.NewStruct(new(models.Group)).For(d.Flavor)
	selectBuilder := groupStruct.SelectFrom(d.Flavor.Quote("groups"))
	selectBuilder.Where(selectBuilder.Equal("group_identifier", groupIdentifier))
	return d.getGroupCommon(tx, selectBuilder, groupStruct)
}

func (d *CommonDB) getGroupCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, groupStruct *sqlbuilder.Struct) (*models.Group, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var group models.Group
	if rows.Next() {
		addr := groupStruct.Addr(&group)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan group")
		}
		return &group, nil
	}

	return nil, nil
}
