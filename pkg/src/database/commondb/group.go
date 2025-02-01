package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
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

func (d *CommonDB) UpdateGroup(tx *sql.Tx, group *models.Group) error {
	if group.Id == 0 {
		return errors.WithStack(errors.New("can't update group with id 0"))
	}

	originalUpdatedAt := group.UpdatedAt
	group.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	groupStruct := sqlbuilder.NewStruct(new(models.Group)).For(d.Flavor)
	updateBuilder := groupStruct.WithoutTag("pk").WithoutTag("dont-update").Update(d.Flavor.Quote("groups"), group)
	updateBuilder.Where(updateBuilder.Equal("id", group.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		group.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update group")
	}

	return nil
}

func (d *CommonDB) CountGroupMembers(tx *sql.Tx, groupId int64) (count int, err error) {
	if groupId <= 0 {
		return 0, errors.WithStack(errors.New("group id must be greater than 0"))
	}

	selectBuilder := d.Flavor.NewSelectBuilder()
	selectBuilder.Select("count(*)").From("users_groups")
	selectBuilder.Where(selectBuilder.Equal("group_id", groupId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return 0, errors.Wrap(err, "unable to scan count")
		}
		return
	}

	return 0, nil
}

func (d *CommonDB) DeleteGroup(tx *sql.Tx, groupId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.Group)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom(d.Flavor.Quote("groups"))
	deleteBuilder.Where(deleteBuilder.Equal("id", groupId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete group")
	}

	return nil
}

func (d *CommonDB) GroupsLoadAttributes(tx *sql.Tx, groups []models.Group) error {
	if groups == nil {
		return nil
	}

	groupIds := make([]int64, len(groups))
	for i, group := range groups {
		groupIds[i] = group.Id
	}

	groupAttributes, err := d.GetGroupAttributesByGroupIds(tx, groupIds)
	if err != nil {
		return errors.Wrap(err, "unable to get group attributes")
	}

	groupAttributesMap := make(map[int64][]models.GroupAttribute)
	for _, groupAttribute := range groupAttributes {
		groupAttributesMap[groupAttribute.GroupId] = append(groupAttributesMap[groupAttribute.GroupId], groupAttribute)
	}

	for i, group := range groups {
		group.Attributes = groupAttributesMap[group.Id]
		groups[i] = group
	}

	return nil
}

func (d *CommonDB) GroupLoadPermissions(tx *sql.Tx, group *models.Group) error {
	if group == nil {
		return nil
	}

	groupPermissions, err := d.GetGroupPermissionsByGroupIds(tx, []int64{group.Id})
	if err != nil {
		return errors.Wrap(err, "unable to get group permissions")
	}

	permissionIds := make([]int64, len(groupPermissions))
	for i, groupPermission := range groupPermissions {
		permissionIds[i] = groupPermission.PermissionId
	}

	permissions, err := d.GetPermissionsByIds(tx, permissionIds)
	if err != nil {
		return errors.Wrap(err, "unable to get permissions")
	}

	group.Permissions = make([]models.Permission, len(permissions))
	copy(group.Permissions, permissions)

	return nil
}

func (d *CommonDB) GroupsLoadPermissions(tx *sql.Tx, groups []models.Group) error {
	if groups == nil {
		return nil
	}

	groupIds := make([]int64, len(groups))
	for i, group := range groups {
		groupIds[i] = group.Id
	}

	groupPermissions, err := d.GetGroupPermissionsByGroupIds(tx, groupIds)
	if err != nil {
		return errors.Wrap(err, "unable to get group permissions")
	}

	permissionIds := make([]int64, len(groupPermissions))
	for i, groupPermission := range groupPermissions {
		permissionIds[i] = groupPermission.PermissionId
	}

	permissions, err := d.GetPermissionsByIds(tx, permissionIds)
	if err != nil {
		return errors.Wrap(err, "unable to get permissions")
	}

	permissionsMap := make(map[int64]models.Permission)
	for _, permission := range permissions {
		permissionsMap[permission.Id] = permission
	}

	groupPermissionsMap := make(map[int64][]models.GroupPermission)
	for _, groupPermission := range groupPermissions {
		groupPermissionsMap[groupPermission.GroupId] = append(groupPermissionsMap[groupPermission.GroupId], groupPermission)
	}

	for i, group := range groups {
		group.Permissions = make([]models.Permission, len(groupPermissionsMap[group.Id]))
		for j, groupPermission := range groupPermissionsMap[group.Id] {
			group.Permissions[j] = permissionsMap[groupPermission.PermissionId]
		}
		groups[i] = group
	}

	return nil
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
