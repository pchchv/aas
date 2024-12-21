package commondb

import (
	"database/sql"

	"github.com/huandu/go-sqlbuilder"
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
