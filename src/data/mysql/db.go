package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/data/commondb"
)

type MySQLDB struct {
	DB       *sql.DB
	CommonDB *commondb.CommonDB
}
