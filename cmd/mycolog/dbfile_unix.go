//go:build !windows
// +build !windows

package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func getDBFilename() (string, error) {
	dbDir := os.Getenv("XDG_DATA_HOME")
	if dbDir == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return "", errors.New("could not find a place for the database")
		}
		dbDir = filepath.Join(home, ".local", "share")
	}
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		log.Printf("Creating directory '%s' for the database file.\n", dbDir)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return "", err
		}
	}
	return filepath.Join(dbDir, "mycolog.sqlite3"), nil
}
