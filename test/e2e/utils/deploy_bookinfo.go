package utils

import (
	"io/ioutil"
	"log"
)

func DeployBookInfo(namespace string) error {
	return KubectlApply(namespace, IstioBookinfoYaml)
}

func DeployBookInfoWithInject(namespace, istioNamespace string) error {
	injected, err := IstioInject(istioNamespace, IstioBookinfoYaml)
	if err != nil {
		return err
	}

	return KubectlApply(namespace, injected)
}

// loads bookinfo from <root>/test/e2e/istio/files/bookinfo.yaml
var IstioBookinfoYaml = func() string {
	bookinfoYamlFile := MustTestFile("bookinfo.yaml")
	b, err := ioutil.ReadFile(bookinfoYamlFile)
	if err != nil {
		log.Fatalf("failed to read bookinfo for test: %v", err)
	}
	return string(b)
}()
