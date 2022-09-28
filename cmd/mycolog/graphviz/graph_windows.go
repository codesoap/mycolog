//go:build windows
// +build windows

package graphviz

import (
	"log"
	"os"
	"path/filepath"
)

// init tries to find the path of the Graphviz binaries on windows and
// adds it to PATH, if found. These paths would match the search:
// - `C:\Program Files (x86)\Graphviz2.38\bin`
// - `E:\Program Files (x86)\Graphviz2.38\bin`
// - `C:\Program Files\Graphviz2.38\bin`
// - `C:\Program Files\Graphviz3.42\bin`
func init() {
	var matches []string
	var err error
	for _, drive := range "CDEFGHIJKLMNOPQRSTUVWXYZ" {
		matches, err = filepath.Glob(string(drive) + `:\Program Files*\Graphviz*\bin`)
		if err != nil {
			panic(err)
		}
		if len(matches) > 0 {
			break
		}
	}
	if len(matches) == 0 {
		log.Println(`No Graphviz installation found.`)
		log.Println(`Graphs will not be available.`)
		return
	}
	os.Setenv("PATH", matches[0]+";"+os.Getenv("PATH"))
}
