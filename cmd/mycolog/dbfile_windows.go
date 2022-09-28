//go:build windows
// +build windows

package main

import (
	"os"
	"path/filepath"
)

func getDBFilename() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "mycolog.sqlite3"), nil
}
