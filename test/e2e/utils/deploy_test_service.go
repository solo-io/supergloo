package utils

import (
	"io/ioutil"
	"log"
)

func DeployTestService(namespace string) error {
	return KubectlApply(namespace, TestServiceYaml)
}

// loads test-service from <root>/test/e2e/files/test-service.yaml
var TestServiceYaml = func() string {
	file := MustTestFile("test-service.yaml")
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read file for test: %v", err)
	}
	return string(b)
}()
