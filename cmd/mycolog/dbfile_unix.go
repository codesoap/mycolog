//go:build !windows
// +build !windows

package main

import (
	"errors"
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
	return filepath.Join(dbDir, "mycolog.sqlite3"), nil
}
