package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateUserAttribute(tx *sql.Tx, userAttribute *models.UserAttribute) error {
	if userAttribute.UserId == 0 {
		return errors.WithStack(errors.New("can't create userAttribute with user_id 0"))
	}

	now := time.Now().UTC()
	originalCreatedAt := userAttribute.CreatedAt
	originalUpdatedAt := userAttribute.UpdatedAt
	userAttribute.CreatedAt = sql.NullTime{Time: now, Valid: true}
	userAttribute.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	userAttributeStruct := sqlbuilder.NewStruct(new(models.UserAttribute)).For(d.Flavor)
	insertBuilder := userAttributeStruct.WithoutTag("pk").InsertInto("user_attributes", userAttribute)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		userAttribute.CreatedAt = originalCreatedAt
		userAttribute.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert userAttribute")
	}

	id, err := result.LastInsertId()
	if err != nil {
		userAttribute.CreatedAt = originalCreatedAt
		userAttribute.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	userAttribute.Id = id
	return nil
}

func (d *CommonDB) GetUserAttributesByUserId(tx *sql.Tx, userId int64) (userAttributes []models.UserAttribute, err error) {
	userAttributeStruct := sqlbuilder.NewStruct(new(models.UserAttribute)).For(d.Flavor)
	selectBuilder := userAttributeStruct.SelectFrom("user_attributes")
	selectBuilder.Where(selectBuilder.Equal("user_id", userId))
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	for rows.Next() {
		var userAttribute models.UserAttribute
		addr := userAttributeStruct.Addr(&userAttribute)
		err = rows.Scan(addr...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan userAttribute")
		}
		userAttributes = append(userAttributes, userAttribute)
	}

	return
}

func (d *CommonDB) GetUserAttributeById(tx *sql.Tx, userAttributeId int64) (*models.UserAttribute, error) {
	userAttributeStruct := sqlbuilder.NewStruct(new(models.UserAttribute)).For(d.Flavor)
	selectBuilder := userAttributeStruct.SelectFrom("user_attributes")
	selectBuilder.Where(selectBuilder.Equal("id", userAttributeId))
	return d.getUserAttributeCommon(tx, selectBuilder, userAttributeStruct)
}

func (d *CommonDB) UpdateUserAttribute(tx *sql.Tx, userAttribute *models.UserAttribute) error {
	if userAttribute.Id == 0 {
		return errors.WithStack(errors.New("can't update userAttribute with id 0"))
	}

	originalUpdatedAt := userAttribute.UpdatedAt
	userAttribute.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	userAttributeStruct := sqlbuilder.NewStruct(new(models.UserAttribute)).For(d.Flavor)
	updateBuilder := userAttributeStruct.WithoutTag("pk").WithoutTag("dont-update").Update("user_attributes", userAttribute)
	updateBuilder.Where(updateBuilder.Equal("id", userAttribute.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		userAttribute.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update userAttribute")
	}

	return nil
}

func (d *CommonDB) DeleteUserAttribute(tx *sql.Tx, userAttributeId int64) error {
	clientStruct := sqlbuilder.NewStruct(new(models.UserAttribute)).For(d.Flavor)
	deleteBuilder := clientStruct.DeleteFrom("user_attributes")
	deleteBuilder.Where(deleteBuilder.Equal("id", userAttributeId))
	sql, args := deleteBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		return errors.Wrap(err, "unable to delete userAttribute")
	}

	return nil
}

func (d *CommonDB) getUserAttributeCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, userAttributeStruct *sqlbuilder.Struct) (*models.UserAttribute, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var userAttribute models.UserAttribute
	if rows.Next() {
		addr := userAttributeStruct.Addr(&userAttribute)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan userAttribute")
		}
		return &userAttribute, nil
	}

	return nil, nil
}
