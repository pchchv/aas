package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
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
