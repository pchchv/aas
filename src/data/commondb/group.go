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
