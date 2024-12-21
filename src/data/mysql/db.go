package mysqldb

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/huandu/go-sqlbuilder"
	"github.com/pchchv/aas/src/config"
	"github.com/pchchv/aas/src/data/commondb"
	"github.com/pkg/errors"
)

type MySQLDB struct {
	DB       *sql.DB
	CommonDB *commondb.CommonDB
}

func NewMySQLDB() (*MySQLDB, error) {
	slog.Info("using database mysql")
	slog.Info(fmt.Sprintf("db username: %v", config.GetDatabase().Username))
	slog.Info(fmt.Sprintf("db host: %v", config.GetDatabase().Host))
	slog.Info(fmt.Sprintf("db port: %v", config.GetDatabase().Port))
	slog.Info(fmt.Sprintf("db name: %v", config.GetDatabase().Name))

	dsnWithoutDBname := fmt.Sprintf("%v:%v@tcp(%v:%v)/?charset=utf8mb4&parseTime=True&loc=UTC",
		config.GetDatabase().Username,
		config.GetDatabase().Password,
		config.GetDatabase().Host,
		config.GetDatabase().Port)
	dsnWithDBname := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=UTC&multiStatements=true",
		config.GetDatabase().Username,
		config.GetDatabase().Password,
		config.GetDatabase().Host,
		config.GetDatabase().Port,
		config.GetDatabase().Name)

	db, err := sql.Open("mysql", dsnWithoutDBname)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open database")
	}

	// create the database if it does not exist
	createDatabaseCommand := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %v CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;", config.GetDatabase().Name)
	if _, err = db.Exec(createDatabaseCommand); err != nil {
		return nil, errors.Wrap(err, "unable to create database")
	}

	db, err = sql.Open("mysql", dsnWithDBname)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open database")
	}

	commonDb := commondb.NewCommonDB(db, sqlbuilder.MySQL)
	mysqlDb := MySQLDB{
		DB:       db,
		CommonDB: commonDb,
	}

	return &mysqlDb, nil
}

func (d *MySQLDB) BeginTransaction() (*sql.Tx, error) {
	return d.CommonDB.BeginTransaction()
}

func (d *MySQLDB) CommitTransaction(tx *sql.Tx) error {
	return d.CommonDB.CommitTransaction(tx)
}
