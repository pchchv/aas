package sessionstore

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/src/database"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
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

func (store *SQLStore) load(session *sessions.Session) (err error) {
	sessIDint, err := parseSessionID(session.ID)
	if err != nil {
		return
	}

	var sess *models.HttpSession
	if sess, err = store.db.GetHttpSessionById(nil, sessIDint); err != nil {
		return
	} else if sess == nil {
		return errors.WithStack(errors.New("session not found"))
	}

	if time.Until(sess.ExpiresOn.Time) < 0 {
		return errors.WithStack(errors.New("session expired"))
	}

	if err = securecookie.DecodeMulti(session.Name(), sess.Data, &session.Values, store.Codecs...); err != nil {
		return
	}

	session.Values["created_on"] = sess.CreatedAt.Time
	session.Values["modified_on"] = sess.UpdatedAt.Time
	session.Values["expires_on"] = sess.ExpiresOn.Time
	return nil
}

func (store *SQLStore) insert(session *sessions.Session) (err error) {
	var createdOn time.Time
	var modifiedOn time.Time
	var expiresOn time.Time
	now := time.Now().UTC()
	if crOn := session.Values["created_on"]; crOn == nil {
		createdOn = now
	} else {
		createdOn = crOn.(time.Time)
	}

	modifiedOn = createdOn
	if exOn := session.Values["expires_on"]; exOn == nil {
		expiresOn = now.Add(time.Second * time.Duration(session.Options.MaxAge))
	} else {
		expiresOn = exOn.(time.Time)
	}

	delete(session.Values, "created_on")
	delete(session.Values, "expires_on")
	delete(session.Values, "modified_on")

	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, store.Codecs...)
	if err != nil {
		return
	}

	sess := models.HttpSession{
		Data:      encoded,
		CreatedAt: sql.NullTime{Time: createdOn, Valid: true},
		UpdatedAt: sql.NullTime{Time: modifiedOn, Valid: true},
		ExpiresOn: sql.NullTime{Time: expiresOn, Valid: true},
	}

	if err = store.db.CreateHttpSession(nil, &sess); err != nil {
		return
	}

	session.ID = fmt.Sprintf("%d", sess.Id)
	return nil
}

func (store *SQLStore) save(session *sessions.Session) error {
	if session.IsNew {
		return store.insert(session)
	}

	var createdOn time.Time
	var expiresOn time.Time
	now := time.Now().UTC()
	if crOn := session.Values["created_on"]; crOn == nil {
		createdOn = now
	} else {
		createdOn = crOn.(time.Time)
	}

	if exOn := session.Values["expires_on"]; exOn == nil {
		expiresOn = now.Add(time.Second * time.Duration(session.Options.MaxAge))
	} else {
		expiresOn = exOn.(time.Time)
		if expiresOn.Sub(now.Add(time.Second*time.Duration(session.Options.MaxAge))) < 0 {
			expiresOn = now.Add(time.Second * time.Duration(session.Options.MaxAge))
		}
	}

	delete(session.Values, "created_on")
	delete(session.Values, "expires_on")
	delete(session.Values, "modified_on")
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, store.Codecs...)
	if err != nil {
		return err
	}

	sessIDint, err := parseSessionID(session.ID)
	if err != nil {
		return err
	}

	sess := models.HttpSession{
		Id:        sessIDint,
		Data:      encoded,
		CreatedAt: sql.NullTime{Time: createdOn, Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
		ExpiresOn: sql.NullTime{Time: expiresOn, Valid: true},
	}

	return store.db.UpdateHttpSession(nil, &sess)
}

// deleteExpired deletes expired sessions from the database.
func (store *SQLStore) deleteExpired() error {
	return store.db.DeleteHttpSessionExpired(nil)
}

func parseSessionID(sessionID string) (sessIDint int64, err error) {
	if n, err := fmt.Sscanf(sessionID, "%d", &sessIDint); err != nil {
		return 0, errors.Wrapf(err, "unable to parse session ID: %s", sessionID)
	} else if n != 1 {
		return 0, errors.WithStack(fmt.Errorf("unable to parse session ID: %s", sessionID))
	}

	return sessIDint, nil
}
