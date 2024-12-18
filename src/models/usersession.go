package models

import (
	"database/sql"
	"time"
)

type UserSession struct {
	Id                         int64               `db:"id" fieldtag:"pk"`
	User                       User                `db:"-"`
	UserId                     int64               `db:"user_id"`
	Started                    time.Time           `db:"started"`
	Clients                    []UserSessionClient `db:"-"`
	AcrLevel                   string              `db:"acr_level"`
	AuthTime                   time.Time           `db:"auth_time"`
	DeviceOS                   string              `db:"device_os"`
	IpAddress                  string              `db:"ip_address"`
	CreatedAt                  sql.NullTime        `db:"created_at" fieldtag:"dont-update"`
	UpdatedAt                  sql.NullTime        `db:"updated_at"`
	DeviceName                 string              `db:"device_name"`
	DeviceType                 string              `db:"device_type"`
	AuthMethods                string              `db:"auth_methods"`
	LastAccessed               time.Time           `db:"last_accessed"`
	SessionIdentifier          string              `db:"session_identifier"`
	Level2AuthConfigHasChanged bool                `db:"level2_auth_config_has_changed"`
}

func (us *UserSession) isValidSinceStarted(userSessionMaxLifetimeInSeconds int) bool {
	utcNow := time.Now().UTC()
	max := us.Started.Add(time.Second * time.Duration(userSessionMaxLifetimeInSeconds))
	return utcNow.Before(max) || utcNow.Equal(max)
}

func (us *UserSession) isValidSinceLastAcessed(userSessionIdleTimeoutInSeconds int) bool {
	utcNow := time.Now().UTC()
	max := us.LastAccessed.Add(time.Second * time.Duration(userSessionIdleTimeoutInSeconds))
	return utcNow.Before(max) || utcNow.Equal(max)
}
