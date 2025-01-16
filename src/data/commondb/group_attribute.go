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
