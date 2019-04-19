package utils

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
)

// returns absolute path of project root
func ProjectRoot() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get runtime.Caller")
	}
	thisDir := filepath.Dir(thisFile)

	return filepath.Abs(filepath.Join(thisDir, "..", "..", ".."))
}

func MustProjectRoot() string {
	r, err := ProjectRoot()
	if err != nil {
		log.Fatalf("failed to get project root: %v", err)
	}
	return r
}

func MustTestFiles() string {
	return filepath.Join(MustProjectRoot(), "test", "e2e", "files")
}

func MustTestFile(filename string) string {
	return filepath.Join(MustTestFiles(), filename)
}
