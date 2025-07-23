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
	dataDir, err := getDataDir()
	return filepath.Join(dataDir, "mycolog.sqlite3"), err
}

func getDataDir() (string, error) {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return "", errors.New("could not find a place for the database")
		}
		dataDir = filepath.Join(home, ".local", "share")
	}
	dataDir = filepath.Join(dataDir, "mycolog")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.Printf("Creating data directory '%s'.\n", dataDir)
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return "", err
		}
	}
	return dataDir, nil
}

func migrateDBFileTo(dbFilename string) error {
	dbFilenameV1, err := getDBFilenameV1()
	if err != nil {
		return err
	}
	// Ignore error; it's OK if there is nothing to migrate:
	_ = os.Rename(dbFilenameV1, dbFilename)
	return nil
}

// getDBFilenameV1 returns the filename of the first and now deprecated
// path of the database.
func getDBFilenameV1() (string, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return "", err
	}
	oldDataDir := filepath.Dir(dataDir) // Trim "/mycolog".
	return filepath.Join(oldDataDir, "mycolog.sqlite3"), nil
}
