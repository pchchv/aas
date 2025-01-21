package sqlitedb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *SQLiteDB) CreateSettings(tx *sql.Tx, settings *models.Settings) error {
	return d.CommonDB.CreateSettings(tx, settings)
}

func (d *SQLiteDB) UpdateSettings(tx *sql.Tx, settings *models.Settings) error {
	return d.CommonDB.UpdateSettings(tx, settings)
}

func (d *SQLiteDB) GetSettingsById(tx *sql.Tx, settingsId int64) (*models.Settings, error) {
	return d.CommonDB.GetSettingsById(tx, settingsId)
}
