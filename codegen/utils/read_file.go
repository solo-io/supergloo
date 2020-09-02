package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// reads a file. path is relative to the file where the caller lies
func MustReadFile(relativePath string) string {
	_, file, _, _ := runtime.Caller(1)
	// chdir to the file who called us
	dir := filepath.Dir(file)
	if err := os.Chdir(dir); err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadFile(filepath.Join(dir, relativePath))
	if err != nil {
		log.Fatal(err)
	}
	return string(b)

}
