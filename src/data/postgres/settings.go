package postgresdb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *PostgresDB) CreateSettings(tx *sql.Tx, settings *models.Settings) error {
	now := time.Now().UTC()
	originalCreatedAt := settings.CreatedAt
	originalUpdatedAt := settings.UpdatedAt
	settings.CreatedAt = sql.NullTime{Time: now, Valid: true}
	settings.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	settingsStruct := sqlbuilder.NewStruct(new(models.Settings)).For(sqlbuilder.PostgreSQL)
	insertBuilder := settingsStruct.WithoutTag("pk").InsertInto("settings", settings)
	sql, args := insertBuilder.Build()
	sql += " RETURNING id"
	rows, err := d.CommonDB.QuerySql(tx, sql, args...)
	if err != nil {
		settings.CreatedAt = originalCreatedAt
		settings.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert settings")
	}
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&settings.Id); err != nil {
			settings.CreatedAt = originalCreatedAt
			settings.UpdatedAt = originalUpdatedAt
			return errors.Wrap(err, "unable to scan settings id")
		}
	}

	return nil
}

func (d *PostgresDB) UpdateSettings(tx *sql.Tx, settings *models.Settings) error {
	return d.CommonDB.UpdateSettings(tx, settings)
}

func (d *PostgresDB) GetSettingsById(tx *sql.Tx, settingsId int64) (*models.Settings, error) {
	return d.CommonDB.GetSettingsById(tx, settingsId)
}
