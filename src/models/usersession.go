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
