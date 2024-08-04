package store

import (
	"database/sql"
	"fmt"
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
	if err = mycoDB.updateDB(); err != nil {
		return mycoDB, err
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

// updateDB updates an existing database to the latest schema version.
func (db *DB) updateDB() error {
	version, err := db.schemaVersion()
	if err != nil || version == 1 {
		return err
	} else if version > 1 {
		return fmt.Errorf("unknown schema version '%d'", version)
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	newVersion := 0
	switch version {
	case 0:
		for _, c := range schemaV1 {
			if _, err = tx.Exec(c); err != nil {
				return err
			}
		}
		newVersion = 1
	}
	_, err = tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, newVersion))
	if err != nil {
		format := "could not update schema version to '%d': %v"
		return fmt.Errorf(format, newVersion, err)
	}
	return tx.Commit()
}

func (db *DB) schemaVersion() (int, error) {
	rows, err := db.Query(`PRAGMA user_version`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	version := 0
	versionFound := false
	for rows.Next() {
		if versionFound {
			return 0, fmt.Errorf("found multiple database schema versions")
		}
		if err := rows.Scan(&version); err != nil {
			return 0, err
		}
		versionFound = true
	}
	if !versionFound {
		return 0, fmt.Errorf("found no database schema version")
	}
	return version, nil
}
