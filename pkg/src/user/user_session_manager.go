package user

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/pchchv/aas/pkg/src/constants"
	"github.com/pchchv/aas/pkg/src/database"
	"github.com/pchchv/aas/pkg/src/models"
	"github.com/pchchv/aas/pkg/src/oauth"
	"github.com/pchchv/aas/pkg/src/useragent"
	"github.com/pkg/errors"
)

type UserSessionManager struct {
	codeIssuer   *oauth.CodeIssuer
	sessionStore sessions.Store
	database     database.Database
}

func NewUserSessionManager(codeIssuer *oauth.CodeIssuer, sessionStore sessions.Store, database database.Database) *UserSessionManager {
	return &UserSessionManager{
		codeIssuer:   codeIssuer,
		sessionStore: sessionStore,
		database:     database,
	}
}

func (u *UserSessionManager) HasValidUserSession(ctx context.Context, userSession *models.UserSession, requestedMaxAgeInSeconds *int) bool {
	settings := ctx.Value(constants.ContextKeySettings).(*models.Settings)
	if userSession != nil {
		return userSession.IsValid(settings.UserSessionIdleTimeoutInSeconds, settings.UserSessionMaxLifetimeInSeconds, requestedMaxAgeInSeconds)
	}

	return false
}

func (u *UserSessionManager) StartNewUserSession(w http.ResponseWriter, r *http.Request, userId int64, clientId int64, authMethods string, acrLevel string) (*models.UserSession, error) {
	utcNow := time.Now().UTC()
	ipWithoutPort, _, _ := net.SplitHostPort(r.RemoteAddr)
	if len(ipWithoutPort) == 0 {
		ipWithoutPort = r.RemoteAddr
	}

	userSession := &models.UserSession{
		SessionIdentifier: uuid.New().String(),
		Started:           utcNow,
		LastAccessed:      utcNow,
		IpAddress:         ipWithoutPort,
		AuthMethods:       authMethods,
		AcrLevel:          acrLevel,
		AuthTime:          utcNow,
		UserId:            userId,
		DeviceName:        useragent.GetDeviceName(r),
		DeviceType:        useragent.GetDeviceType(r),
		DeviceOS:          useragent.GetDeviceOS(r),
	}

	userSession.Clients = append(userSession.Clients, models.UserSessionClient{
		Started:      utcNow,
		LastAccessed: utcNow,
		ClientId:     clientId,
	})

	tx, err := u.database.BeginTransaction()
	if err != nil {
		return nil, err
	}
	defer u.database.RollbackTransaction(tx) //nolint:errcheck

	if err = u.database.CreateUserSession(tx, userSession); err != nil {
		return nil, err
	}

	for _, client := range userSession.Clients {
		client.UserSessionId = userSession.Id
		if err = u.database.CreateUserSessionClient(tx, &client); err != nil {
			return nil, err
		}
	}

	if err = u.database.CommitTransaction(tx); err != nil {
		return nil, err
	}

	allUserSessions, err := u.database.GetUserSessionsByUserId(nil, userId)
	if err != nil {
		return nil, err
	}

	// delete other sessions from this same device & ip
	for _, us := range allUserSessions {
		if us.SessionIdentifier != userSession.SessionIdentifier &&
			us.DeviceName == userSession.DeviceName &&
			us.DeviceType == userSession.DeviceType &&
			us.DeviceOS == userSession.DeviceOS &&
			us.IpAddress == ipWithoutPort {
			err = u.database.DeleteUserSession(nil, us.Id)
			if err != nil {
				return nil, err
			}
		}
	}

	sess, err := u.sessionStore.Get(r, constants.SessionName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get the session")
	}

	sess.Values[constants.SessionKeySessionIdentifier] = userSession.SessionIdentifier
	if err = u.sessionStore.Save(r, w, sess); err != nil {
		return nil, err
	}

	return userSession, nil
}

func (u *UserSessionManager) BumpUserSession(r *http.Request, sessionIdentifier string, clientId int64) (*models.UserSession, error) {
	userSession, err := u.database.GetUserSessionBySessionIdentifier(nil, sessionIdentifier)
	if err != nil {
		return nil, err
	}

	if userSession != nil {
		if err = u.database.UserSessionLoadClients(nil, userSession); err != nil {
			return nil, err
		}

		utcNow := time.Now().UTC()
		userSession.LastAccessed = utcNow
		// concatenate any new IP address
		ipWithoutPort, _, _ := net.SplitHostPort(r.RemoteAddr)
		if len(ipWithoutPort) == 0 {
			ipWithoutPort = r.RemoteAddr
		}

		if !strings.Contains(userSession.IpAddress, ipWithoutPort) {
			userSession.IpAddress = fmt.Sprintf("%v,%v", userSession.IpAddress, ipWithoutPort)
		}

		// append client if not already present
		clientFound := false
		for _, c := range userSession.Clients {
			if c.ClientId == clientId {
				clientFound = true
				break
			}
		}

		if clientFound {
			// update last accessed
			for i, c := range userSession.Clients {
				if c.ClientId == clientId {
					userSession.Clients[i].LastAccessed = utcNow
					break
				}
			}
		} else {
			userSession.Clients = append(userSession.Clients, models.UserSessionClient{
				Started:      utcNow,
				LastAccessed: utcNow,
				ClientId:     clientId,
			})
		}

		tx, err := u.database.BeginTransaction()
		if err != nil {
			return nil, err
		}
		defer u.database.RollbackTransaction(tx) //nolint:errcheck

		if err = u.database.UpdateUserSession(tx, userSession); err != nil {
			return nil, err
		}

		for _, client := range userSession.Clients {
			if client.Id > 0 {
				if err = u.database.UpdateUserSessionClient(tx, &client); err != nil {
					return nil, err
				}
			} else {
				// insert new
				client.UserSessionId = userSession.Id
				if err = u.database.CreateUserSessionClient(tx, &client); err != nil {
					return nil, err
				}
			}
		}

		if err = u.database.CommitTransaction(tx); err != nil {
			return nil, err
		}

		return userSession, nil
	}

	return nil, errors.WithStack(errors.New("Unexpected: can't bump user session because user session is nil"))
}
