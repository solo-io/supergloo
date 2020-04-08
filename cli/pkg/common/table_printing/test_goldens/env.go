package test_goldens

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	UPDATE_GOLDENS  = "UPDATE_GOLDENS"
	GoldenDirectory = "test_goldens"
)

func UpdateGoldens() bool {
	return os.Getenv(UPDATE_GOLDENS) != ""
}

func GoldenFilePath(dir, filename string) string {
	return filepath.Join(GoldenDirectory, dir, fmt.Sprintf("%s.golden", filename))
}
