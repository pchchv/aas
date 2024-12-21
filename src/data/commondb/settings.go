package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) GetSettingsById(tx *sql.Tx, settingsId int64) (*models.Settings, error) {
	settingsStruct := sqlbuilder.NewStruct(new(models.Settings)).For(d.Flavor)
	selectBuilder := settingsStruct.SelectFrom("settings")
	selectBuilder.Where(selectBuilder.Equal("id", settingsId))

	return d.getSettingsCommon(tx, selectBuilder, settingsStruct)
}

func (d *CommonDB) CreateSettings(tx *sql.Tx, settings *models.Settings) error {
	now := time.Now().UTC()
	originalCreatedAt := settings.CreatedAt
	originalUpdatedAt := settings.UpdatedAt
	settings.CreatedAt = sql.NullTime{Time: now, Valid: true}
	settings.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	settingsStruct := sqlbuilder.NewStruct(new(models.Settings)).For(d.Flavor)
	insertBuilder := settingsStruct.WithoutTag("pk").InsertInto("settings", settings)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		settings.CreatedAt = originalCreatedAt
		settings.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert settings")
	}

	id, err := result.LastInsertId()
	if err != nil {
		settings.CreatedAt = originalCreatedAt
		settings.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	settings.Id = id
	return nil
}

func (d *CommonDB) getSettingsCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, settingsStruct *sqlbuilder.Struct) (*models.Settings, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var settings models.Settings
	if rows.Next() {
		addr := settingsStruct.Addr(&settings)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan settings")
		}
		return &settings, nil
	}

	return nil, nil
}
