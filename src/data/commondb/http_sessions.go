package commondb

import (
	"database/sql"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) GetHttpSessionById(tx *sql.Tx, httpSessionId int64) (*models.HttpSession, error) {
	httpSessionStruct := sqlbuilder.NewStruct(new(models.HttpSession)).For(d.Flavor)
	selectBuilder := httpSessionStruct.SelectFrom("http_sessions")
	selectBuilder.Where(selectBuilder.Equal("id", httpSessionId))
	return d.getHttpSessionCommon(tx, selectBuilder, httpSessionStruct)
}

func (d *CommonDB) getHttpSessionCommon(tx *sql.Tx, selectBuilder *sqlbuilder.SelectBuilder, httpSessionStruct *sqlbuilder.Struct) (*models.HttpSession, error) {
	sql, args := selectBuilder.Build()
	rows, err := d.QuerySql(tx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query database")
	}
	defer rows.Close()

	var httpSession models.HttpSession
	if rows.Next() {
		addr := httpSessionStruct.Addr(&httpSession)
		if err = rows.Scan(addr...); err != nil {
			return nil, errors.Wrap(err, "unable to scan httpSession")
		}
		return &httpSession, nil
	}

	return nil, nil
}
