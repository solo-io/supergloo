package manifests

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

var (
	//go:embed operator/* traffic/*
	manifestFiles embed.FS
)

func RenderOperator(operatorFile string, data interface{}) (string, error) {
	filePath := "operator/" + operatorFile
	file, err := manifestFiles.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed finding file %s template: %w", filePath, err)
	}
	fileTemplate, err := template.New(operatorFile).Parse(string(file))
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
	file, err := manifestFiles.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed finding file %s template: %w", filePath, err)
	}
	fileTemplate, err := template.New(operatorFile).Parse(string(file))
	if err != nil {
		return "", fmt.Errorf("failed preparing %q template: %w", operatorFile, err)
	}

	b := new(strings.Builder)
	if err := fileTemplate.Execute(b, data); err != nil {
		return "", fmt.Errorf("failed rendering %q: %w", operatorFile, err)
	}
	return b.String(), nil
}
