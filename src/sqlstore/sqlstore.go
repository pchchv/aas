package sessionstore

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/src/database"
)

type SQLStore struct {
	db      database.Database
	Codecs  []securecookie.Codec
	Options *sessions.Options
}

func NewSQLStore(db database.Database, path string, maxAge int, httpOnly bool, secure bool, sameSite http.SameSite, keyPairs ...[]byte) *SQLStore {
	codecs := securecookie.CodecsFromPairs(keyPairs...)
	for _, codec := range codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxLength(1024 * 64) // 64k
		}
	}

	return &SQLStore{
		db:     db,
		Codecs: codecs,
		Options: &sessions.Options{
			Path:     path,
			MaxAge:   maxAge,
			HttpOnly: httpOnly,
			Secure:   secure,
			SameSite: sameSite,
		},
	}
}

func init() {
	gob.Register(time.Time{})
}
