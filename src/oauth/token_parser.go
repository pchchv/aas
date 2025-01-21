package oauth

import db "github.com/pchchv/aas/src/data"

type TokenParser struct {
	database db.Database
}

func NewTokenParser(database db.Database) *TokenParser {
	return &TokenParser{
		database: database,
	}
}
