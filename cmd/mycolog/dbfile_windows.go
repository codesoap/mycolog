//go:build windows
// +build windows

package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
)

func getDBFilename() (string, error) {
	dataDir, err := getDataDir()
	return filepath.Join(dataDir, "mycolog.sqlite3"), err
}

func getDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dataDir := path.Join(homeDir, "mycolog")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.Printf("Creating data directory '%s'.\n", dataDir)
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return "", err
		}
	}
	return filepath.Join(homeDir, "mycolog"), nil
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "mycolog.sqlite3"), nil
}
