package sqlitedb

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"strings"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/config"
	"github.com/pchchv/aas/src/database/commondb"
	"github.com/pkg/errors"
	sqlitedriver "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var sqliteMigrationsFs embed.FS

type SQLiteDB struct {
	DB       *sql.DB
	CommonDB *commondb.CommonDB
}

func NewSQLiteDB() (*SQLiteDB, error) {
	dsn := config.GetDatabase().DSN
	if dsn == "" {
		dsn = "file::memory:?cache=shared"
	}

	slog.Info("using database sqlite")
	slog.Info(fmt.Sprintf("db dsn: %v", config.GetDatabase().DSN))

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open database")
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	// Execute PRAGMA statements directly
	pragmaStatements := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
	}

	// Only set journal_mode to WAL if it's not an in-memory database
	isMemoryDB := strings.Contains(dsn, ":memory:")
	if !isMemoryDB {
		pragmaStatements = append(pragmaStatements, "PRAGMA journal_mode = WAL;")
	}

	for _, stmt := range pragmaStatements {
		if _, err = db.Exec(stmt); err != nil {
			return nil, errors.Wrapf(err, "failed to execute %s", stmt)
		}
	}

	// Verify PRAGMA settings
	pragmaChecks := []struct {
		name     string
		query    string
		expected interface{}
	}{
		{"foreign_keys", "PRAGMA foreign_keys;", 1},
		{"busy_timeout", "PRAGMA busy_timeout;", 5000},
	}

	// Only check journal_mode if it's not an in-memory database
	if !isMemoryDB {
		pragmaChecks = append(pragmaChecks, struct {
			name     string
			query    string
			expected interface{}
		}{"journal_mode", "PRAGMA journal_mode;", "wal"})
	}

	for _, check := range pragmaChecks {
		var value interface{}
		if err = db.QueryRow(check.query).Scan(&value); err != nil {
			return nil, errors.Wrapf(err, "unable to check %s status", check.name)
		}

		if value.(string) != check.expected.(string) {
			return nil, errors.Errorf("%s is not set correctly. Expected %v, got %v", check.name, check.expected, value)
		}
	}

	if err := db.PingContext(context.Background()); err != nil {
		if errWithCode, ok := err.(*sqlitedriver.Error); ok {
			err = errors.WithStack(errors.New(sqlitedriver.ErrorCodeString[errWithCode.Code()]))
		}
		return nil, errors.WithStack(fmt.Errorf("sqlite ping: %w", err))
	}

	slog.Info("connected to sqlite database with required PRAGMA settings")
	commonDb := commondb.NewCommonDB(db, sqlbuilder.SQLite)
	sqliteDb := SQLiteDB{
		DB:       db,
		CommonDB: commonDb,
	}

	return &sqliteDb, nil
}

func (d *SQLiteDB) BeginTransaction() (*sql.Tx, error) {
	return d.CommonDB.BeginTransaction()
}

func (d *SQLiteDB) CommitTransaction(tx *sql.Tx) error {
	return d.CommonDB.CommitTransaction(tx)
}

func (d *SQLiteDB) RollbackTransaction(tx *sql.Tx) error {
	return d.CommonDB.RollbackTransaction(tx)
}

func (d *SQLiteDB) Migrate() error {
	driver, err := sqlite.WithInstance(d.DB, &sqlite.Config{})
	if err != nil {
		return errors.Wrap(err, "unable to create migration driver")
	}

	iofs, err := iofs.New(sqliteMigrationsFs, "migrations")
	if err != nil {
		return errors.Wrap(err, "unable to create migration filesystem")
	}

	migrate, err := gomigrate.NewWithInstance("iofs", iofs, "sqlite", driver)
	if err != nil {
		return errors.Wrap(err, "unable to create migration instance")
	}

	if err = migrate.Up(); err != nil && err != gomigrate.ErrNoChange {
		return errors.Wrap(err, "unable to migrate the database")
	} else if err != nil && err == gomigrate.ErrNoChange {
		slog.Info("no need to migrate the database")
	}

	return nil
}

func (d *SQLiteDB) IsEmpty() (bool, error) {
	return d.CommonDB.IsEmpty()
}
