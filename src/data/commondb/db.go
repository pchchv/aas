package commondb

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/config"
	"github.com/pkg/errors"
)

type CommonDB struct {
	DB     *sql.DB
	Flavor sqlbuilder.Flavor
}

func NewCommonDB(db *sql.DB, flavor sqlbuilder.Flavor) *CommonDB {
	return &CommonDB{
		DB:     db,
		Flavor: flavor,
	}
}

func (d *CommonDB) Log(sql string, args ...any) {
	if config.Get().LogSQL {
		var argsStr string
		slog.Info(fmt.Sprintf("sql: %v", sql))
		for i, arg := range args {
			argsStr += fmt.Sprintf("[arg %v: %v] ", i, arg)
		}
		slog.Info(fmt.Sprintf("sql args: %v", argsStr))
	}
}

func (d *CommonDB) ExecSql(tx *sql.Tx, sql string, args ...any) (result sql.Result, err error) {
	d.Log(sql, args...)
	if tx != nil {
		if result, err = tx.Exec(sql, args...); err != nil {
			result, err = nil, errors.Wrap(err, "unable to execute SQL")
		}
		return
	}

	if result, err = d.DB.Exec(sql, args...); err != nil {
		result, err = nil, errors.Wrap(err, "unable to execute SQL")
	}
	return
}

func (d *CommonDB) QuerySql(tx *sql.Tx, sql string, args ...any) (result *sql.Rows, err error) {
	d.Log(sql, args...)
	if tx != nil {
		if result, err = tx.Query(sql, args...); err != nil {
			result, err = nil, errors.Wrap(err, "unable to execute SQL")
		}
		return
	}

	if result, err = d.DB.Query(sql, args...); err != nil {
		result, err = nil, errors.Wrap(err, "unable to execute SQL")
	}
	return
}

func (d *CommonDB) BeginTransaction() (tx *sql.Tx, err error) {
	if config.Get().LogSQL {
		slog.Info("beginning transaction")
	}

	if tx, err = d.DB.Begin(); err != nil {
		return nil, errors.Wrap(err, "unable to begin transaction")
	}
	return
}

func (d *CommonDB) RollbackTransaction(tx *sql.Tx) (err error) {
	if config.Get().LogSQL {
		slog.Info("rolling back transaction")
	}

	if err = tx.Rollback(); err != nil {
		return errors.Wrap(err, "unable to rollback transaction")
	}

	return nil
}
