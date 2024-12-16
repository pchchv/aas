package enums

const (
	TokenTypeId TokenType = iota
	TokenTypeBearer
	TokenTypeRefresh

	AcrLevel1          AcrLevel = "urn:goiabada:level1"
	AcrLevel2Optional  AcrLevel = "urn:goiabada:level2_optional"
	AcrLevel2Mandatory AcrLevel = "urn:goiabada:level2_mandatory"
)

type TokenType int

func (tt TokenType) String() string {
	return []string{"ID", "Bearer", "Refresh"}[tt]
}

type AcrLevel string

func (acrl AcrLevel) String() string {
	return string(acrl)
}
