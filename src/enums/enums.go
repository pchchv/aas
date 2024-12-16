package enums

const (
	TokenTypeId TokenType = iota
	TokenTypeBearer
	TokenTypeRefresh
)

type TokenType int

func (tt TokenType) String() string {
	return []string{"ID", "Bearer", "Refresh"}[tt]
}
