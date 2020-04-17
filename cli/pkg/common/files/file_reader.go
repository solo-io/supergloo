package files

import (
	"io/ioutil"
	"os"
)

//go:generate mockgen -destination ./mocks/mock_file_reader.go -source ./file_reader.go

type FileReader interface {
	Read(filePath string) ([]byte, error)
	Exists(filePath string) (exists bool, err error)
}

func NewDefaultFileReader() FileReader {
	return &fileReader{}
}

type fileReader struct{}

func (f *fileReader) Read(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func (f *fileReader) Exists(filePath string) (exists bool, err error) {
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
