package timezones

import "time"

type Zone struct {
	*time.Location
	CountryCode string
	Zone        string
	Abbr        []string
	CountryName string
	Comments    string
}
