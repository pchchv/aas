package commondb

import (
	"database/sql"

	"github.com/huandu/go-sqlbuilder"
)

type CommonDB struct {
	DB     *sql.DB
	Flavor sqlbuilder.Flavor
}
