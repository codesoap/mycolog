package store

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const timeFormat = "2006-01-02"

// DB is the database used to store all persistent data.
type DB struct {
	*sql.DB
}

// GetDB returns the database. If it does not exist, it will be created
// and initialized first.
func GetDB(filepath string) (DB, error) {
	newDB := false
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		newDB = true
	}
	db, err := sql.Open("sqlite3", filepath)
	mycoDB := DB{db}
	if err != nil {
		return mycoDB, err
	}
	db.SetMaxOpenConns(1) // See github.com/mattn/go-sqlite3/issues/274
	if newDB {
		if err = mycoDB.initDB(); err != nil {
			return mycoDB, err
		}
	}
	_, err = db.Exec(`PRAGMA foreign_keys = ON`)
	return mycoDB, err
}

// initDB configures the database and creates all tables.
func (db *DB) initDB() (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	for _, c := range schemaV0 {
		if _, err = tx.Exec(c); err != nil {
			return
		}
	}
	return tx.Commit()
}
