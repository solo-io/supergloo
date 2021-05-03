package packr

import (
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"strings"
	"text/template"
)

const (
	examplesPath = "../manifests"
)

var box = packr.New("files", examplesPath)

func RenderOperator(operatorFile string, data interface{}) (string, error) {
	filePath := "operator/" + operatorFile
	file, err := box.FindString(filePath)
	if err != nil {
		return "", fmt.Errorf("failed finding file %s template: %w", filePath, err)
	}
	fileTemplate, err := template.New(operatorFile).Parse(file)
	if err != nil {
		return "", fmt.Errorf("failed preparing %q template: %w", operatorFile, err)
	}

	b := new(strings.Builder)
	if err := fileTemplate.Execute(b, data); err != nil {
		return "", fmt.Errorf("failed rendering %q: %w", operatorFile, err)
	}
	return b.String(), nil
}

func RenderTestFile(operatorFile string, data interface{}) (string, error) {
	filePath := "traffic/" + operatorFile
	file, err := box.FindString(filePath)
	if err != nil {
		return "", fmt.Errorf("failed finding file %s template: %w", filePath, err)
	}
	fileTemplate, err := template.New(operatorFile).Parse(file)
	if err != nil {
		return "", fmt.Errorf("failed preparing %q template: %w", operatorFile, err)
	}

	b := new(strings.Builder)
	if err := fileTemplate.Execute(b, data); err != nil {
		return "", fmt.Errorf("failed rendering %q: %w", operatorFile, err)
	}
	return b.String(), nil
}
