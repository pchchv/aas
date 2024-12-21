package commondb

import (
	"database/sql"

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
