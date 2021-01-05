package global

import (
	"os"
	"path/filepath"
)

func WorkDir() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
	return dir
}
