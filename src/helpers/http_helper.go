package helpers

import (
	"io/fs"

	"github.com/pchchv/aas/src/database"
)

type HttpHelper struct {
	templateFS fs.FS
	database   database.Database
}

func NewHttpHelper(templateFS fs.FS, database database.Database) *HttpHelper {
	return &HttpHelper{
		templateFS: templateFS,
		database:   database,
	}
}
