package gencheck

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

// This is a utility for ensuring that all generated files are generated as expected
// Usage:
// 1. make a directory with files and filenames that match what you expect your tests to generate
// 2. create a new FileGenerationObjective from that directory
// 3. run your tests, calling GetFileContent once per test-generated content
// 4. after all tests are run, call AllFilesVisited to verify that all of the tests expected files were generated at least once

type FileGenerationObjective struct {
	// hold a map of all the files that we expect to generate in this test
	targetFiles map[string]string
	// keep track of all the files (by name) that we generate
	// after the suite, make sure that all expected files have been generated at least once
	visitedFiles map[string]int
	// all files that we expect to generate are held in the this directory
	referenceFileDir string
	lock             sync.Mutex
}

func NewFileGenerationObjective(parentDir string) (*FileGenerationObjective, error) {
	targetFiles := make(map[string]string)
	files, err := ioutil.ReadDir(parentDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			return nil, fmt.Errorf("found dir %v in reference directory %v - only files are supported",
				file.Name(),
				parentDir)
		}
		content, err := ioutil.ReadFile(filepath.Join(parentDir, file.Name()))
		if err != nil {
			return nil, err
		}
		targetFiles[file.Name()] = string(content)
	}
	return &FileGenerationObjective{
		targetFiles: targetFiles,
		// this will be populated during the test run
		visitedFiles:     make(map[string]int),
		referenceFileDir: parentDir,
	}, nil

}

func (fgo *FileGenerationObjective) markVisited(filename string) {
	fgo.lock.Lock()
	defer fgo.lock.Unlock()
	if count, exists := fgo.visitedFiles[filename]; exists {
		fgo.visitedFiles[filename] = count + 1
	} else {
		fgo.visitedFiles[filename] = 1
	}
}

func (fgo *FileGenerationObjective) GetFileContent(filename string) (string, error) {
	fgo.markVisited(filename)
	referenceContent, exists := fgo.targetFiles[filename]
	if !exists {
		return "", fmt.Errorf("file %v not found in reference dir %v",
			filename,
			fgo.referenceFileDir)
	}
	return referenceContent, nil
}

func (fgo *FileGenerationObjective) HasFile(filename, expectedContent string) (bool, error) {
	referenceContent, err := fgo.GetFileContent(filename)
	if err != nil {
		return false, err
	}
	if referenceContent != expectedContent {
		return false, fmt.Errorf("content of file %v in reference dir %v does not match expected:\n%v",
			filename,
			fgo.referenceFileDir,
			expectedContent)
	}
	return true, nil
}

// Recommendation: call this after all generated files have been tested to ensure that you have covered all cases
func (fgo *FileGenerationObjective) AllFilesVisited() bool {
	if len(fgo.visitedFiles) == len(fgo.targetFiles) {
		return true
	}
	return false
}
