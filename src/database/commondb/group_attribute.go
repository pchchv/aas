package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateGroupAttribute(tx *sql.Tx, groupAttribute *models.GroupAttribute) error {
	if groupAttribute.GroupId == 0 {
		return errors.WithStack(errors.New("can't create groupAttribute with group_id 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := groupAttribute.CreatedAt
	originalUpdatedAt := groupAttribute.UpdatedAt
	groupAttribute.CreatedAt = sql.NullTime{Time: now, Valid: true}
	groupAttribute.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	groupAttributeStruct := sqlbuilder.NewStruct(new(models.GroupAttribute)).For(d.Flavor)
	insertBuilder := groupAttributeStruct.WithoutTag("pk").InsertInto("group_attributes", groupAttribute)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		groupAttribute.CreatedAt = originalCreatedAt
		groupAttribute.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert groupAttribute")
	}

	id, err := result.LastInsertId()
	if err != nil {
		groupAttribute.CreatedAt = originalCreatedAt
		groupAttribute.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	groupAttribute.Id = id
	return nil
}

func (d *CommonDB) GetGroupAttributesByGroupIds(tx *sql.Tx, groupIds []int64) (groupAttributes []models.GroupAttribute, err error) {
	if len(groupIds) == 0 {
		return nil, nil
	}

	groupAttributeStruct := sqlbuilder.NewStruct(new(models.GroupAttribute)).For(d.Flavor)
	selectBuilder := groupAttributeStruct.SelectFrom("group_attributes")
	selectBuilder.Where(selectBuilder.In("group_id", sqlbuilder.Flatten(groupIds)...))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var groupAttribute models.GroupAttribute
		addr := groupAttributeStruct.Addr(&groupAttribute)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan groupAttribute")
		}
		groupAttributes = append(groupAttributes, groupAttribute)
	}

	return
}

func (d *CommonDB) GetGroupAttributesByGroupId(tx *sql.Tx, groupId int64) (groupAttributes []models.GroupAttribute, err error) {
	groupAttributeStruct := sqlbuilder.NewStruct(new(models.GroupAttribute)).For(d.Flavor)
	selectBuilder := groupAttributeStruct.SelectFrom("group_attributes")
	selectBuilder.Where(selectBuilder.Equal("group_id", groupId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var groupAttribute models.GroupAttribute
		addr := groupAttributeStruct.Addr(&groupAttribute)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan groupAttribute")
		}
		groupAttributes = append(groupAttributes, groupAttribute)
	}

	return
}

func (d *CommonDB) GetGroupAttributeById(tx *sql.Tx, groupAttributeId int64) (*models.GroupAttribute, error) {
	groupAttributeStruct := sqlbuilder.NewStruct(new(models.GroupAttribute)).For(d.Flavor)
	selectBuilder := groupAttributeStruct.SelectFrom("group_attributes")
	selectBuilder.Where(selectBuilder.Equal("id", groupAttributeId))
	return d.getGroupAttributeCommon(tx, selectBuilder, groupAttributeStruct)
}

func (d *CommonDB) UpdateGroupAttribute(tx *sql.Tx, groupAttribute *models.GroupAttribute) error {
	if groupAttribute.Id == 0 {
		return errors.WithStack(errors.New("can't update groupAttribute with id 0"))
	}

	originalUpdatedAt := groupAttribute.UpdatedAt
	groupAttribute.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	groupAttributeStruct := sqlbuilder.NewStruct(new(models.GroupAttribute)).For(d.Flavor)
	updateBuilder := groupAttributeStruct.WithoutTag("pk").WithoutTag("dont-update").Update("group_attributes", groupAttribute)
	updateBuilder.Where(updateBuilder.Equal("id", groupAttribute.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		groupAttribute.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update groupAttribute")
	}

	return nil
}

func (d *CommonDB) DeleteGroupAttribute(tx *sql.Tx, groupAttributeId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.GroupAttribute)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("group_attributes")
	deleteBuilder.Where(deleteBuilder.Equal("id", groupAttributeId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete groupAttribute")
	}

	return nil
}

func (d *CommonDB) getGroupAttributeCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, groupAttributeStruct *sqlbuilder.Struct) (*models.GroupAttribute, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var groupAttribute models.GroupAttribute
	if rows.Next() {
		addr := groupAttributeStruct.Addr(&groupAttribute)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan groupAttribute")
		}
		return &groupAttribute, nil
	}

	return nil, nil
}
