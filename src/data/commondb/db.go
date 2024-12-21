package commondb

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/config"
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
