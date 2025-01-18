package commondb

import (
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/models"
	"github.com/pkg/errors"
)

func (d *CommonDB) CreateKeyPair(tx *sql.Tx, keyPair *models.KeyPair) error {
	now := time.Now().UTC()
	originalCreatedAt := keyPair.CreatedAt
	originalUpdatedAt := keyPair.UpdatedAt
	keyPair.CreatedAt = sql.NullTime{Time: now, Valid: true}
	keyPair.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	keyPairStruct := sqlbuilder.NewStruct(new(models.KeyPair)).For(d.Flavor)
	insertBuilder := keyPairStruct.WithoutTag("pk").InsertInto("key_pairs", keyPair)
	sql, args := insertBuilder.Build()
	result, err := d.ExecSql(tx, sql, args...)
	if err != nil {
		keyPair.CreatedAt = originalCreatedAt
		keyPair.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to insert keyPair")
	}

	id, err := result.LastInsertId()
	if err != nil {
		keyPair.CreatedAt = originalCreatedAt
		keyPair.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to get last insert id")
	}

	keyPair.Id = id
	return nil
}

func (d *CommonDB) UpdateKeyPair(tx *sql.Tx, keyPair *models.KeyPair) error {
	if keyPair.Id == 0 {
		return errors.WithStack(errors.New("can't update keyPair with id 0"))
	}

	originalUpdatedAt := keyPair.UpdatedAt
	keyPair.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	keyPairStruct := sqlbuilder.NewStruct(new(models.KeyPair)).For(d.Flavor)
	updateBuilder := keyPairStruct.WithoutTag("pk").WithoutTag("dont-update").Update("key_pairs", keyPair)
	updateBuilder.Where(updateBuilder.Equal("id", keyPair.Id))
	sql, args := updateBuilder.Build()
	if _, err := d.ExecSql(tx, sql, args...); err != nil {
		keyPair.UpdatedAt = originalUpdatedAt
		return errors.Wrap(err, "unable to update keyPair")
	}

	return nil
}
